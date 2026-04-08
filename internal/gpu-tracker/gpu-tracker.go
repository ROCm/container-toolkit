/**
# Copyright (c) Advanced Micro Devices, Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package gpuTracker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/gofrs/flock"
)

// Accessibility represents the access mode of a GPU
type Accessibility string

const (
	SHARED_ACCESS    Accessibility = "Shared"
	EXCLUSIVE_ACCESS Accessibility = "Exclusive"
)

// accessibility is the internal int representation used for JSON serialization
// of the tracker file, kept for backward compatibility with existing tracker files.
type accessibility int

const (
	sharedAccessInt accessibility = iota
	exclusiveAccessInt
)

func (a accessibility) toAccessibility() (Accessibility, error) {
	switch a {
	case sharedAccessInt:
		return SHARED_ACCESS, nil
	case exclusiveAccessInt:
		return EXCLUSIVE_ACCESS, nil
	default:
		return "", fmt.Errorf("invalid accessibility value: %d", int(a))
	}
}

// GPUStatusEntry represents the status of a single GPU
type GPUStatusEntry struct {
	GPUId         int
	UUID          string
	Accessibility Accessibility
	ContainerIds  []string
}

// AccessibilityResult contains the outcome of a MakeGPUsExclusive or MakeGPUsShared operation
type AccessibilityResult struct {
	Changed       []int
	NotChanged    []int
	InvalidGPUs   []string
	InvalidRanges []string
}

// Interface for GPU Tracker package
type Interface interface {
	// Initialize GPU Tracker
	Init() error

	// Check if GPU Tracker is enabled
	IsEnabled() (bool, error)

	// Enable GPU Tracker
	Enable() error

	// Disable GPU Tracker
	Disable() error

	// Reset GPU Tracker
	Reset() error

	// Show GPUs Status
	ShowStatus() ([]GPUStatusEntry, error)

	// Make specified GPUs exclusive such that they can be used
	// by at most one container at any instance
	MakeGPUsExclusive(gpus string) (*AccessibilityResult, error)

	// Make specified GPUs shared such that they can be used
	// by any number of containers at any instance
	MakeGPUsShared(gpus string) (*AccessibilityResult, error)

	// Reserve GPUs for a container
	ReserveGPUs(gpus string, containerId string) ([]int, error)

	// Release all GPUs linked to a container
	ReleaseGPUs(containerId string) error
}

type gpu_status_t struct {
	// UUID of GPU
	UUID string `json:"uuid"`

	// Partition Type of the GPU
	PartitionType string `json:"partitionType"`

	// GPU accessibility (int for backward-compatible JSON serialization)
	Accessibility accessibility `json:"accessibility"`

	// Container Ids of the containers to which the GPU is assigned
	ContainerIds []string `json:"containerIds"`
}

type gpu_tracker_data_t struct {
	// Status of GPU Tracker
	Enabled bool `json:"enabled"`

	// Status of all GPUs
	GPUsStatus map[int]gpu_status_t `json:"gpusStatus"`

	// Info of all GPUs
	GPUsInfo map[int]amdgpu.DeviceInfo `json:"gpusInfo"`
}

// isGPUTrackerInitializedTYpe is the type for functions
// that return if GPU Tracker is initialized
type isGPUTrackerInitializedType func() (bool, error)

// initializeGPUTrackerType is the type for functions that
// initialize GPU Tracker
type initializeGPUTrackerType func() error

// parseGPUsListType is the type for functions that parse
// GPU list strings and returns the valid and invalid GPU Ids
type parseGPUsListType func(string) ([]int, []string, []string, error)

// readGPUTrackerFileType is the type for functions that
// read the GPU Tracker file and return the GPUs status
type readGPUTrackerFileType func() (gpu_tracker_data_t, error)

// writeGPUTrackerFileType is the type for functions that
// write the GPUs status to GPU Tracker file
type writeGPUTrackerFileType func(gpu_tracker_data_t) error

// validateGPUsInfoType is the type for functions that
// validate the GPUs info
type validateGPUsInfoType func(map[int]amdgpu.DeviceInfo) (bool, error)

type gpu_tracker_t struct {
	// path to GPU Tracker lock file
	gpuTrackerLockFile string

	// function to check if GPU Tracker is initialized
	isGPUTrackerInitialized isGPUTrackerInitializedType

	// function to initialize GPU Tracker
	initializeGPUTracker initializeGPUTrackerType

	// function to parse GPU list strings
	parseGPUsList parseGPUsListType

	// function to read GPU Tracker file
	readGPUTrackerFile readGPUTrackerFileType

	// function to write GPU Tracker file
	writeGPUTrackerFile writeGPUTrackerFileType

	// function to validate GPUs info
	validateGPUsInfo validateGPUsInfoType
}

const (
	gpuTrackerFile     = "/var/log/gpu-tracker.json"
	gpuTrackerLockFile = "/var/log/gpu-tracker.lock"
)

const defaultLockTimeout = 10 * time.Second

func acquireLock(lockFile string, timeout time.Duration) (*flock.Flock, error) {
	lock := flock.New(lockFile)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	locked, err := lock.TryLockContext(ctx, 100*time.Millisecond)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("acquiring lock: timeout exceeded")
		}
		return nil, fmt.Errorf("acquiring lock: %w", err)
	}
	if !locked {
		return nil, fmt.Errorf("acquiring lock: timeout exceeded")
	}

	return lock, nil
}

func parseGPUsList(gpus string) ([]int, []string, []string, error) {
	// isHexString checks if a string contains only hexadecimal characters
	isHexString := func(s string) bool {
		if len(s) == 0 {
			return false
		}
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}

	validGPUs := []int{}
	invalidGPUs := []string{}
	invalidGPUsRange := []string{}

	gpusInfo, err := amdgpu.GetAMDGPUs()
	if err != nil {
		return []int{}, []string{}, []string{}, fmt.Errorf("getting AMD GPU info: %w", err)
	}

	if gpus == "all" || gpus == "All" || gpus == "ALL" {
		for i := 0; i < len(gpusInfo); i++ {
			validGPUs = append(validGPUs, i)
		}
		return validGPUs, []string{}, []string{}, nil
	}

	uuidToGPUIdMap, err := amdgpu.GetUniqueIdToDeviceIndexMap()
	if err != nil {
		uuidToGPUIdMap = make(map[string][]int) // Continue with empty map
	}

	for _, c := range strings.Split(gpus, ",") {
		if strings.HasPrefix(c, "0x") || strings.HasPrefix(c, "0X") ||
			(len(c) > 8 && isHexString(c)) {
			uuid := strings.ToLower(c)
			if !strings.HasPrefix(uuid, "0x") {
				uuid = "0x" + uuid
			}
			if gpuIds, exists := uuidToGPUIdMap[uuid]; exists {
				validGPUs = append(validGPUs, gpuIds...)
			} else {
				uuid = strings.TrimPrefix(uuid, "0x")
				if gpuIds, exists := uuidToGPUIdMap[uuid]; exists {
					validGPUs = append(validGPUs, gpuIds...)
				} else {
					invalidGPUs = append(invalidGPUs, c)
				}
			}
		} else if strings.Contains(c, "-") {
			devsRange := strings.SplitN(c, "-", 2)
			start, err0 := strconv.Atoi(devsRange[0])
			end, err1 := strconv.Atoi(devsRange[1])
			if err0 != nil || err1 != nil ||
				start < 0 || end < 0 || start > end {
				invalidGPUsRange = append(invalidGPUsRange, c)
			} else {
				for i := start; i <= end; i++ {
					if i < len(gpusInfo) {
						validGPUs = append(validGPUs, i)
					} else {
						invalidGPUs = append(invalidGPUs, strconv.Itoa(i))
					}
				}
			}
		} else {
			i, err := strconv.Atoi(c)
			if err == nil {
				if i >= 0 && i < len(gpusInfo) {
					validGPUs = append(validGPUs, i)
				} else {
					invalidGPUs = append(invalidGPUs, c)
				}
			} else {
				invalidGPUs = append(invalidGPUs, c)
			}
		}
	}

	sort.Ints(validGPUs)

	return validGPUs, invalidGPUs, invalidGPUsRange, nil
}

func isGPUTrackerInitialized() (bool, error) {
	gpuTrackerInitialized := false
	_, err := os.Stat(gpuTrackerFile)
	if err == nil {
		gpuTrackerInitialized = true
	} else {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("checking file %v: %w", gpuTrackerFile, err)
		}
	}

	return gpuTrackerInitialized, nil
}

func readGPUTrackerFile() (gpu_tracker_data_t, error) {
	file, err := os.Open(gpuTrackerFile)
	if err != nil {
		return gpu_tracker_data_t{GPUsStatus: make(map[int]gpu_status_t), GPUsInfo: make(map[int]amdgpu.DeviceInfo)},
			fmt.Errorf("opening GPU tracker file: %w", err)
	}
	defer file.Close()

	var gpuTrackerData gpu_tracker_data_t
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&gpuTrackerData); err != nil {
		return gpu_tracker_data_t{GPUsStatus: make(map[int]gpu_status_t), GPUsInfo: make(map[int]amdgpu.DeviceInfo)},
			fmt.Errorf("decoding GPU tracker JSON: %w", err)
	}

	return gpuTrackerData, nil
}

func writeGPUTrackerFile(gpuTrackerData gpu_tracker_data_t) error {
	tempPath := gpuTrackerFile + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	encoder := json.NewEncoder(tempFile)
	if err := encoder.Encode(gpuTrackerData); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("encoding JSON to temp file: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("syncing temp file: %w", err)
	}

	tempFile.Close()

	if err := os.Rename(tempPath, gpuTrackerFile); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

func initializeGPUTracker() error {
	gpusInfo, err := amdgpu.GetAMDGPUs()
	if err != nil {
		return fmt.Errorf("getting AMD GPU info: %w", err)
	}

	uuidToGPUIdMap, err := amdgpu.GetUniqueIdToDeviceIndexMap()
	if err != nil {
		uuidToGPUIdMap = make(map[string][]int) // Continue with empty map
	}
	gpuIdToUUIDMap := make(map[int]string)
	for uuid, gpuIds := range uuidToGPUIdMap {
		if strings.HasPrefix(uuid, "0x") || strings.HasPrefix(uuid, "0X") {
			uuid = uuid[2:]
		}
		uuid = "0x" + strings.ToUpper(uuid)
		for _, gpuId := range gpuIds {
			gpuIdToUUIDMap[gpuId] = uuid
		}
	}

	gpuTrackerData := gpu_tracker_data_t{Enabled: false, GPUsStatus: make(map[int]gpu_status_t), GPUsInfo: make(map[int]amdgpu.DeviceInfo)}
	for gpuId, gpuInfo := range gpusInfo {
		gpuTrackerData.GPUsInfo[gpuId] = gpuInfo
		gpuTrackerData.GPUsStatus[gpuId] = gpu_status_t{
			UUID:          gpuIdToUUIDMap[gpuId],
			PartitionType: gpusInfo[gpuId].PartitionType,
			Accessibility: sharedAccessInt,
			ContainerIds:  []string{},
		}
	}

	return writeGPUTrackerFile(gpuTrackerData)
}

func validateGPUsInfo(savedGPUsInfo map[int]amdgpu.DeviceInfo) (bool, error) {
	tempGPUsInfo, err := amdgpu.GetAMDGPUs()
	if err != nil {
		return false, fmt.Errorf("getting AMD GPU info: %w", err)
	}
	currentGPUsInfo := make(map[int]amdgpu.DeviceInfo)
	for gpuId, gpuInfo := range tempGPUsInfo {
		currentGPUsInfo[gpuId] = gpuInfo
	}

	if !reflect.DeepEqual(savedGPUsInfo, currentGPUsInfo) {
		return false, nil
	}

	return true, nil
}

func (gpuTracker *gpu_tracker_t) IsEnabled() (bool, error) {
	initialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		return false, err
	}
	if !initialized {
		return false, nil
	}

	gpuTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		return false, err
	}

	return gpuTrackerData.Enabled, nil
}

func (gpuTracker *gpu_tracker_t) Init() error {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile, defaultLockTimeout)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	if err = gpuTracker.initializeGPUTracker(); err != nil {
		return err
	}

	logger.Log.Printf("GPU Tracker has been initialized")
	return nil
}

func (gpuTracker *gpu_tracker_t) Enable() error {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile, defaultLockTimeout)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		return err
	}

	if !gpuTrackerInitialized {
		if err := gpuTracker.initializeGPUTracker(); err != nil {
			return err
		}
	}

	gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		return err
	}

	if gpusTrackerData.Enabled {
		return nil
	}

	if err := gpuTracker.initializeGPUTracker(); err != nil {
		return fmt.Errorf("reinitialize GPU tracker: %w", err)
	}

	gpusTrackerData, err = gpuTracker.readGPUTrackerFile()
	if err != nil {
		return err
	}

	gpusTrackerData.Enabled = true

	return gpuTracker.writeGPUTrackerFile(gpusTrackerData)
}

func (gpuTracker *gpu_tracker_t) Disable() error {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile, defaultLockTimeout)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		return err
	}

	if !gpuTrackerInitialized {
		if err := gpuTracker.initializeGPUTracker(); err != nil {
			return err
		}
	} else {
		gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
		if err != nil {
			return err
		}

		gpusTrackerData.Enabled = false

		if err := gpuTracker.writeGPUTrackerFile(gpusTrackerData); err != nil {
			return err
		}
	}

	return nil
}

func (gpuTracker *gpu_tracker_t) Reset() error {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile, defaultLockTimeout)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		return err
	}

	gpuTrackerEnabled := false

	if !gpuTrackerInitialized {
		if err := gpuTracker.initializeGPUTracker(); err != nil {
			return err
		}
	} else {
		gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
		if err != nil {
			return err
		}

		gpuTrackerEnabled = gpusTrackerData.Enabled

		if err := gpuTracker.initializeGPUTracker(); err != nil {
			return fmt.Errorf("reinitialize GPU tracker: %w", err)
		}

		gpusTrackerData, err = gpuTracker.readGPUTrackerFile()
		if err != nil {
			return err
		}

		if gpuTrackerEnabled {
			gpusTrackerData.Enabled = true
			if err := gpuTracker.writeGPUTrackerFile(gpusTrackerData); err != nil {
				return err
			}
		}
	}

	return nil
}

func (gpuTracker *gpu_tracker_t) ShowStatus() ([]GPUStatusEntry, error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile, defaultLockTimeout)
	if err != nil {
		return nil, err
	}
	defer lock.Unlock()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		return nil, err
	}

	if !gpuTrackerInitialized {
		if err := gpuTracker.initializeGPUTracker(); err != nil {
			return nil, err
		}
	}

	gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		return nil, err
	}

	result, err := gpuTracker.validateGPUsInfo(gpusTrackerData.GPUsInfo)
	if err != nil {
		return nil, fmt.Errorf("validate GPU info: %w", err)
	}
	if !result {
		return nil, fmt.Errorf("GPU info mismatch: please reset GPU Tracker")
	}

	var entries []GPUStatusEntry
	for gpuId := 0; gpuId < len(gpusTrackerData.GPUsStatus); gpuId++ {
		acc, err := gpusTrackerData.GPUsStatus[gpuId].Accessibility.toAccessibility()
		if err != nil {
			return nil, fmt.Errorf("GPU %d: %w", gpuId, err)
		}
		entries = append(entries, GPUStatusEntry{
			GPUId:         gpuId,
			UUID:          gpusTrackerData.GPUsStatus[gpuId].UUID,
			Accessibility: acc,
			ContainerIds:  gpusTrackerData.GPUsStatus[gpuId].ContainerIds,
		})
	}

	return entries, nil
}

func (gpuTracker *gpu_tracker_t) MakeGPUsExclusive(gpus string) (*AccessibilityResult, error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile, defaultLockTimeout)
	if err != nil {
		return nil, err
	}
	defer lock.Unlock()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		return nil, err
	}

	if !gpuTrackerInitialized {
		if err := gpuTracker.initializeGPUTracker(); err != nil {
			return nil, err
		}
	}

	gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		return nil, err
	}

	result, err := gpuTracker.validateGPUsInfo(gpusTrackerData.GPUsInfo)
	if err != nil {
		return nil, fmt.Errorf("validate GPU info: %w", err)
	}
	if !result {
		return nil, fmt.Errorf("GPU info mismatch: please reset GPU Tracker")
	}

	validGPUs, invalidGPUs, invalidGPUsRange, err := gpuTracker.parseGPUsList(gpus)
	if err != nil {
		return nil, err
	}

	res := &AccessibilityResult{
		InvalidGPUs:   invalidGPUs,
		InvalidRanges: invalidGPUsRange,
	}

	for _, gpuId := range validGPUs {
		if len(gpusTrackerData.GPUsStatus[gpuId].ContainerIds) < 2 {
			gpusTrackerData.GPUsStatus[gpuId] = gpu_status_t{
				UUID:          gpusTrackerData.GPUsStatus[gpuId].UUID,
				PartitionType: gpusTrackerData.GPUsStatus[gpuId].PartitionType,
				Accessibility: exclusiveAccessInt,
				ContainerIds:  gpusTrackerData.GPUsStatus[gpuId].ContainerIds,
			}
			res.Changed = append(res.Changed, gpuId)
		} else {
			res.NotChanged = append(res.NotChanged, gpuId)
		}
	}

	if err := gpuTracker.writeGPUTrackerFile(gpusTrackerData); err != nil {
		return nil, err
	}

	return res, nil
}

func (gpuTracker *gpu_tracker_t) MakeGPUsShared(gpus string) (*AccessibilityResult, error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile, defaultLockTimeout)
	if err != nil {
		return nil, err
	}
	defer lock.Unlock()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		return nil, err
	}

	if !gpuTrackerInitialized {
		if err := gpuTracker.initializeGPUTracker(); err != nil {
			return nil, err
		}
	}

	gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		return nil, err
	}

	result, err := gpuTracker.validateGPUsInfo(gpusTrackerData.GPUsInfo)
	if err != nil {
		return nil, fmt.Errorf("validate GPU info: %w", err)
	}
	if !result {
		return nil, fmt.Errorf("GPU info mismatch: please reset GPU Tracker")
	}

	validGPUs, invalidGPUs, invalidGPUsRange, err := gpuTracker.parseGPUsList(gpus)
	if err != nil {
		return nil, err
	}

	for _, gpuId := range validGPUs {
		gpusTrackerData.GPUsStatus[gpuId] = gpu_status_t{
			UUID:          gpusTrackerData.GPUsStatus[gpuId].UUID,
			PartitionType: gpusTrackerData.GPUsStatus[gpuId].PartitionType,
			Accessibility: sharedAccessInt,
			ContainerIds:  gpusTrackerData.GPUsStatus[gpuId].ContainerIds,
		}
	}

	if err := gpuTracker.writeGPUTrackerFile(gpusTrackerData); err != nil {
		return nil, err
	}

	return &AccessibilityResult{
		Changed:       validGPUs,
		InvalidGPUs:   invalidGPUs,
		InvalidRanges: invalidGPUsRange,
	}, nil
}

func (gpuTracker *gpu_tracker_t) ReserveGPUs(gpus string, containerId string) ([]int, error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile, defaultLockTimeout)
	if err != nil {
		return nil, err
	}
	defer lock.Unlock()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		return []int{}, err
	}

	if !gpuTrackerInitialized {
		if err := gpuTracker.initializeGPUTracker(); err != nil {
			return []int{}, err
		}
	}

	gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		return []int{}, err
	}

	validGPUs, invalidGPUs, invalidGPUsRange, err := gpuTracker.parseGPUsList(gpus)
	if err != nil {
		return []int{}, err
	}
	if len(invalidGPUsRange) > 0 {
		logger.Log.Printf("Ignoring %v GPUs Ranges as they are invalid", invalidGPUsRange)
	}
	if len(invalidGPUs) > 0 {
		logger.Log.Printf("Ignoring %v GPUs as they are invalid", invalidGPUs)
	}

	if !gpusTrackerData.Enabled {
		logger.Log.Printf("GPU Tracker is disabled")
		return validGPUs, nil
	}

	result, err := gpuTracker.validateGPUsInfo(gpusTrackerData.GPUsInfo)
	if err != nil {
		return []int{}, fmt.Errorf("validate GPU info: %w", err)
	}
	if !result {
		return []int{}, fmt.Errorf("GPU info mismatch: please reset GPU Tracker")
	}

	var allocatedGPUs []int
	var unavailableGPUs []int
	for _, gpuId := range validGPUs {
		if gpusTrackerData.GPUsStatus[gpuId].Accessibility == sharedAccessInt ||
			(gpusTrackerData.GPUsStatus[gpuId].Accessibility == exclusiveAccessInt &&
				len(gpusTrackerData.GPUsStatus[gpuId].ContainerIds) == 0) {
			gpusTrackerData.GPUsStatus[gpuId] = gpu_status_t{
				UUID:          gpusTrackerData.GPUsStatus[gpuId].UUID,
				PartitionType: gpusTrackerData.GPUsStatus[gpuId].PartitionType,
				Accessibility: gpusTrackerData.GPUsStatus[gpuId].Accessibility,
				ContainerIds:  append(gpusTrackerData.GPUsStatus[gpuId].ContainerIds, containerId),
			}
			allocatedGPUs = append(allocatedGPUs, gpuId)
		} else {
			unavailableGPUs = append(unavailableGPUs, gpuId)
		}
	}

	if err := gpuTracker.writeGPUTrackerFile(gpusTrackerData); err != nil {
		return []int{}, err
	}

	if len(allocatedGPUs) > 0 {
		logger.Log.Printf("GPUs %v allocated", allocatedGPUs)
	}
	if len(unavailableGPUs) > 0 {
		return []int{}, fmt.Errorf("GPUs %v are exclusive and already in use", unavailableGPUs)
	}

	return allocatedGPUs, nil
}

func (gpuTracker *gpu_tracker_t) ReleaseGPUs(containerId string) error {
	removeContainerId := func(containerId string, containerIds []string) ([]string, bool) {
		for idx, id := range containerIds {
			if id == containerId {
				return append(containerIds[:idx], containerIds[idx+1:]...), true
			}
		}
		return containerIds, false
	}

	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile, defaultLockTimeout)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		return err
	}

	if gpuTrackerInitialized {
		gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
		if err != nil {
			return err
		}

		var releasedGPUs []int
		for gpuId, _ := range gpusTrackerData.GPUsStatus {
			containerIds, released := removeContainerId(containerId, gpusTrackerData.GPUsStatus[gpuId].ContainerIds)
			if released {
				gpusTrackerData.GPUsStatus[gpuId] = gpu_status_t{
					UUID:          gpusTrackerData.GPUsStatus[gpuId].UUID,
					PartitionType: gpusTrackerData.GPUsStatus[gpuId].PartitionType,
					Accessibility: gpusTrackerData.GPUsStatus[gpuId].Accessibility,
					ContainerIds:  containerIds,
				}
				releasedGPUs = append(releasedGPUs, gpuId)
			}
		}

		if err := gpuTracker.writeGPUTrackerFile(gpusTrackerData); err != nil {
			return err
		}

		logger.Log.Printf("Released GPUs %v used by container %v", releasedGPUs, containerId)
	}

	return nil
}

func New() (Interface, error) {
	gpuTracker := &gpu_tracker_t{
		gpuTrackerLockFile:      gpuTrackerLockFile,
		isGPUTrackerInitialized: isGPUTrackerInitialized,
		initializeGPUTracker:    initializeGPUTracker,
		parseGPUsList:           parseGPUsList,
		readGPUTrackerFile:      readGPUTrackerFile,
		writeGPUTrackerFile:     writeGPUTrackerFile,
		validateGPUsInfo:        validateGPUsInfo,
	}
	return gpuTracker, nil
}
