/**
# Copyright (c) Advanced Micro Devices, Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the \"License\");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an \"AS IS\" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package gpuTracker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/gofrs/flock"
)

type accessibility int

const (
	SHARED_ACCESS accessibility = iota
	EXCLUSIVE_ACCESS
)

// Interface for GPU Tracker package
type Interface interface {
	// Initialize GPU Tracker
	Init() error

	// Enable GPU Tracker
	Enable() error

	// Disable GPU Tracker
	Disable() error

	// Reset GPU Tracker
	Reset() error

	// Show GPUs Status
	ShowStatus() error

	// Make specified GPUs exclusive such that they can be used
	// by at most one container at any instance
	MakeGPUsExclusive(gpus string) error

	// Make specified GPUs shared such that they can be used
	// by any number of containers at any instance
	MakeGPUsShared(gpus string) error

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

	// GPU accessibility
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

func setupSignalHandler(lock *flock.Flock) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-c
		logger.Log.Printf("Received signal: %v. Cleaning up...", sig)
		fmt.Printf("Received signal: %v. Cleaning up...\n", sig)
		if lock != nil {
			_ = lock.Unlock()
		}
		os.Exit(1)
	}()
}

func acquireLock(lockFile string) (*flock.Flock, error) {
	lock := flock.New(lockFile)

	timeout := time.After(10 * time.Second)
	tick := time.Tick(100 * time.Millisecond)

	for {
		select {
		case <-timeout:
			logger.Log.Printf("Acquiring lock timeout exceeded")
			return nil, fmt.Errorf("Acquiring lock timeout exceeded")
		case <-tick:
			locked, err := lock.TryLock()
			if err != nil {
				logger.Log.Printf("Failed to acquire lock, Error: %v", err)
				return nil, fmt.Errorf("Failed to acquire lock, Error: %v", err)
			}
			if locked {
				return lock, nil
			}
		}
	}
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
		logger.Log.Printf("Failed to get AMD GPUs info, Error: %v", err)
		return []int{}, []string{}, []string{}, fmt.Errorf("Failed to get AMD GPUs info, Error: %v", err)
	}

	if gpus == "all" || gpus == "All" || gpus == "ALL" {
		for i := 0; i < len(gpusInfo); i++ {
			validGPUs = append(validGPUs, i)
		}
		return validGPUs, []string{}, []string{}, nil
	}

	uuidToGPUIdMap, err := amdgpu.GetUniqueIdToDeviceIndexMap()
	if err != nil {
		logger.Log.Printf("Failed to get UUID to GPU Id mappings: %v", err)
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
			return false, fmt.Errorf("Error checking file %v, Error:%v", gpuTrackerFile, err)
		}
	}

	return gpuTrackerInitialized, nil
}

func readGPUTrackerFile() (gpu_tracker_data_t, error) {
	file, err := os.Open(gpuTrackerFile)
	if err != nil {
		logger.Log.Printf("Error opening file, Error: %v", err)
		return gpu_tracker_data_t{GPUsStatus: make(map[int]gpu_status_t), GPUsInfo: make(map[int]amdgpu.DeviceInfo)},
			fmt.Errorf("Error opening file, Error: %v", err)
	}
	defer file.Close()

	var gpuTrackerData gpu_tracker_data_t
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&gpuTrackerData); err != nil {
		logger.Log.Printf("Failed to decode JSON, Error: %v", err)
		return gpu_tracker_data_t{GPUsStatus: make(map[int]gpu_status_t), GPUsInfo: make(map[int]amdgpu.DeviceInfo)},
			fmt.Errorf("Failed to decode JSON, Error: %v", err)
	}

	return gpuTrackerData, nil
}

func writeGPUTrackerFile(gpuTrackerData gpu_tracker_data_t) error {
	tempPath := gpuTrackerFile + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		logger.Log.Printf("Error creating temp file, Error: %v", err)
		return fmt.Errorf("Error creating temp file, Error: %v", err)
	}

	encoder := json.NewEncoder(tempFile)
	if err := encoder.Encode(gpuTrackerData); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		logger.Log.Printf("Error encoding JSON to temp file, Error: %v", err)
		return fmt.Errorf("Error encoding JSON to temp file, Error: %v", err)
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		logger.Log.Printf("Error syncing temp file: %v", err)
		return fmt.Errorf("Error syncing temp file: %v", err)
	}

	tempFile.Close()

	if err := os.Rename(tempPath, gpuTrackerFile); err != nil {
		logger.Log.Printf("Error renaming temp file: %v", err)
		return fmt.Errorf("Error renaming temp file: %v", err)
	}

	return nil
}

func initializeGPUTracker() error {
	gpusInfo, err := amdgpu.GetAMDGPUs()
	if err != nil {
		logger.Log.Printf("Failed to get AMD GPUs info, Error: %v", err)
		return fmt.Errorf("Failed to get AMD GPUs info, Error: %v", err)
	}

	uuidToGPUIdMap, err := amdgpu.GetUniqueIdToDeviceIndexMap()
	if err != nil {
		logger.Log.Printf("Failed to get UUID to GPU Id mappings: %v", err)
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
			Accessibility: SHARED_ACCESS,
			ContainerIds:  []string{},
		}
	}

	return writeGPUTrackerFile(gpuTrackerData)
}

func validateGPUsInfo(savedGPUsInfo map[int]amdgpu.DeviceInfo) (bool, error) {
	tempGPUsInfo, err := amdgpu.GetAMDGPUs()
	if err != nil {
		logger.Log.Printf("Failed to get AMD GPUs info, Error: %v", err)
		return false, fmt.Errorf("Failed to get AMD GPUs info, Error: %v", err)
	}
	currentGPUsInfo := make(map[int]amdgpu.DeviceInfo)
	for gpuId, gpuInfo := range tempGPUsInfo {
		currentGPUsInfo[gpuId] = gpuInfo
	}

	equal := reflect.DeepEqual(savedGPUsInfo, currentGPUsInfo)
	if equal != true {
		logger.Log.Printf("GPUs info is invalid. Please reset GPU Tracker.")
		fmt.Printf("GPUs info is invalid. Please reset GPU Tracker.\n")
		return false, nil
	}

	return true, nil
}

func (gpuTracker *gpu_tracker_t) Init() (err error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("Init lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Printf("Recovered in Init: %v", r)
			err = fmt.Errorf("Panic in Init: %v", r)
		}
	}()

	err = gpuTracker.initializeGPUTracker()
	if err != nil {
		logger.Log.Printf("Failed to initialize GPU Tracker, Error: %v", err)
		return fmt.Errorf("Failed to initialize GPU Tracker, Error: %v", err)
	}

	logger.Log.Printf("GPU Tracker has been initialized")
	return nil
}

func (gpuTracker *gpu_tracker_t) Enable() (err error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("Enable lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Printf("Recovered in Enable: %v", r)
			err = fmt.Errorf("Panic in Enable: %v", r)
		}
	}()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		logger.Log.Printf("Failed to check if GPU Tracker is initialized, Error:%v", err)
		fmt.Printf("Failed to check if GPU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if !gpuTrackerInitialized {
		err := gpuTracker.initializeGPUTracker()
		if err != nil {
			return err
		}
	}

	gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		logger.Log.Printf("Failed to show GPU Tracker status, Error: %v", err)
		fmt.Printf("Failed to show GPU Tracker status, Error: %v\n", err)
		return err
	}

	if gpusTrackerData.Enabled {
		logger.Log.Printf("GPU Tracker is already enabled")
		fmt.Printf("GPU Tracker is already enabled\n")
		return nil
	}

	err = gpuTracker.initializeGPUTracker()
	if err != nil {
		logger.Log.Printf("Failed to enable GPU Tracker, Error: %v", err)
		fmt.Printf("Failed to enable GPU Tracker, Error: %v\n", err)
		return err
	}

	gpusTrackerData, err = gpuTracker.readGPUTrackerFile()
	if err != nil {
		logger.Log.Printf("Failed to enable GPU Tracker, Error: %v", err)
		fmt.Printf("Failed to enable GPU Tracker, Error: %v\n", err)
		return err
	}

	gpusTrackerData.Enabled = true

	err = gpuTracker.writeGPUTrackerFile(gpusTrackerData)
	if err != nil {
		logger.Log.Printf("Failed to enable GPU Tracker, Error: %v", err)
		fmt.Printf("Failed to enable GPU Tracker, Error: %v\n", err)
		return err
	}

	logger.Log.Printf("GPU Tracker has been enabled")
	fmt.Printf("GPU Tracker has been enabled\n")
	return nil
}

func (gpuTracker *gpu_tracker_t) Disable() (err error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("Disable lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Printf("Recovered in Disable: %v", r)
			err = fmt.Errorf("Panic in Disable: %v", r)
		}
	}()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		logger.Log.Printf("Failed to check if GPU Tracker is initialized, Error:%v", err)
		fmt.Printf("Failed to check if GPU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if !gpuTrackerInitialized {
		err := gpuTracker.initializeGPUTracker()
		if err != nil {
			logger.Log.Printf("Failed to disable GPU Tracker, Error: %v", err)
			fmt.Printf("Failed to disable GPU Tracker, Error: %v\n", err)
			return err
		}
	} else {
		gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
		if err != nil {
			logger.Log.Printf("Failed to disable GPU Tracker, Error: %v", err)
			fmt.Printf("Failed to disable GPU Tracker, Error: %v\n", err)
			return err
		}

		gpusTrackerData.Enabled = false

		err = gpuTracker.writeGPUTrackerFile(gpusTrackerData)
		if err != nil {
			logger.Log.Printf("Failed to disable GPU Tracker, Error: %v", err)
			fmt.Printf("Failed to disable GPU Tracker, Error: %v\n", err)
			return err
		}
	}

	logger.Log.Printf("GPU Tracker has been disabled")
	fmt.Printf("GPU Tracker has been disabled\n")
	return nil
}

func (gpuTracker *gpu_tracker_t) Reset() (err error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("Reset lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Printf("Recovered in Reset: %v", r)
			err = fmt.Errorf("Panic in Reset: %v", r)
		}
	}()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		logger.Log.Printf("Failed to check if GPU Tracker is initialized, Error:%v", err)
		fmt.Printf("Failed to check if GPU Tracker is initialized, Error:%v\n", err)
		return err
	}

	gpuTrackerEnabled := false

	if !gpuTrackerInitialized {
		err := gpuTracker.initializeGPUTracker()
		if err != nil {
			logger.Log.Printf("Failed to reset GPU Tracker, Error: %v", err)
			fmt.Printf("Failed to reset GPU Tracker, Error: %v\n", err)
			return err
		}
	} else {
		gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
		if err != nil {
			logger.Log.Printf("Failed to reset GPU Tracker, Error: %v", err)
			fmt.Printf("Failed to reset GPU Tracker, Error: %v\n", err)
			return err
		}

		gpuTrackerEnabled = gpusTrackerData.Enabled

		err = gpuTracker.initializeGPUTracker()
		if err != nil {
			logger.Log.Printf("Failed to reset GPU Tracker, Error: %v", err)
			fmt.Printf("Failed to reset GPU Tracker, Error: %v\n", err)
			return err
		}

		gpusTrackerData, err = gpuTracker.readGPUTrackerFile()
		if err != nil {
			logger.Log.Printf("Failed to reset GPU Tracker, Error: %v", err)
			fmt.Printf("Failed to reset GPU Tracker, Error: %v\n", err)
			return err
		}

		if gpuTrackerEnabled == true {
			gpusTrackerData.Enabled = true
			err = gpuTracker.writeGPUTrackerFile(gpusTrackerData)
			if err != nil {
				logger.Log.Printf("Failed to reset GPU Tracker, Error: %v", err)
				fmt.Printf("Failed to reset GPU Tracker, Error: %v\n", err)
				return err
			}
		}
	}

	logger.Log.Printf("GPU Tracker has been reset")
	fmt.Printf("GPU Tracker has been reset\n")
	if gpuTrackerEnabled {
		fmt.Printf("Since GPU Tracker was enabled, it is recommended to stop and restart running containers to get the most accurate GPU Tracker status\n")
	}
	return nil
}

func (gpuTracker *gpu_tracker_t) ShowStatus() (err error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("ShowStatus lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Printf("Recovered in ShowStatus: %v", r)
			err = fmt.Errorf("Panic in ShowStatus: %v", r)
		}
	}()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		logger.Log.Printf("Failed to check if GPU Tracker is initialized, Error:%v", err)
		fmt.Printf("Failed to check if GPU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if !gpuTrackerInitialized {
		err := gpuTracker.initializeGPUTracker()
		if err != nil {
			return err
		}
	}

	gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		logger.Log.Printf("Failed to show GPU Tracker status, Error: %v", err)
		fmt.Printf("Failed to show GPU Tracker status, Error: %v\n", err)
		return err
	}

	if gpusTrackerData.Enabled == false {
		logger.Log.Printf("GPU Tracker is disabled")
		fmt.Printf("GPU Tracker is disabled\n")
		return nil
	}

	result, err := gpuTracker.validateGPUsInfo(gpusTrackerData.GPUsInfo)
	if err != nil || result != true {
		return err
	}

	fmt.Println(strings.Repeat("-", 120))
	fmt.Printf("%-10s%-25s%-20s%-65s\n", "GPU Id", "UUID", "Accessibility", "Container Ids")
	fmt.Println(strings.Repeat("-", 120))
	for gpuId := 0; gpuId < len(gpusTrackerData.GPUsStatus); gpuId++ {
		var accessibility string
		switch gpusTrackerData.GPUsStatus[gpuId].Accessibility {
		case SHARED_ACCESS:
			accessibility = "Shared"
		case EXCLUSIVE_ACCESS:
			accessibility = "Exclusive"
		default:
			fmt.Printf("Invalid accessibility value %v\n", gpusTrackerData.GPUsStatus[gpuId].Accessibility)
			break
		}

		if len(gpusTrackerData.GPUsStatus[gpuId].ContainerIds) > 0 {
			for idx, id := range gpusTrackerData.GPUsStatus[gpuId].ContainerIds {
				if idx == 0 {
					fmt.Printf("%-10v%-25v%-20v%-65v\n", gpuId, gpusTrackerData.GPUsStatus[gpuId].UUID, accessibility, id)
				} else {
					fmt.Printf("%-10v%-25v%-20v%-65v\n", "", "", "", id)
				}
			}
		} else {
			fmt.Printf("%-10v%-25v%-20v%-65v\n", gpuId, gpusTrackerData.GPUsStatus[gpuId].UUID, accessibility, "-")
		}
	}

	return nil
}

func (gpuTracker *gpu_tracker_t) MakeGPUsExclusive(gpus string) (err error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("MakeGPUsExclusive lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Printf("Recovered in MakeGPUsExclusive: %v", r)
			err = fmt.Errorf("Panic in MakeGPUsExclusive: %v", r)
		}
	}()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		logger.Log.Printf("Failed to check if GPU Tracker is initialized, Error:%v", err)
		fmt.Printf("Failed to check if GPU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if !gpuTrackerInitialized {
		err = gpuTracker.initializeGPUTracker()
		if err != nil {
			return err
		}
	}

	gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		logger.Log.Printf("Failed to make GPUs %v exclusive, Error: %v", gpus, err)
		fmt.Printf("Failed to make GPUs %v exclusive, Error: %v\n", gpus, err)
		return err
	}

	if gpusTrackerData.Enabled == false {
		logger.Log.Printf("GPU Tracker is disabled")
		fmt.Printf("GPU Tracker is disabled\n")
		return nil
	}

	result, err := gpuTracker.validateGPUsInfo(gpusTrackerData.GPUsInfo)
	if err != nil || result != true {
		return err
	}

	validGPUs, invalidGPUs, invalidGPUsRange, err := gpuTracker.parseGPUsList(gpus)
	if err != nil {
		logger.Log.Printf("Failed to parse GPUs list %v, Error: %v", gpus, err)
		fmt.Printf("Failed to parse GPUs list %v, Error: %v\n", gpus, err)
		return err
	}

	gpusMadeExclusive := []int{}
	gpusNotMadeExclusive := []int{}

	for _, gpuId := range validGPUs {
		if len(gpusTrackerData.GPUsStatus[gpuId].ContainerIds) < 2 {
			gpusTrackerData.GPUsStatus[gpuId] = gpu_status_t{
				UUID:          gpusTrackerData.GPUsStatus[gpuId].UUID,
				PartitionType: gpusTrackerData.GPUsStatus[gpuId].PartitionType,
				Accessibility: EXCLUSIVE_ACCESS,
				ContainerIds:  gpusTrackerData.GPUsStatus[gpuId].ContainerIds,
			}
			gpusMadeExclusive = append(gpusMadeExclusive, gpuId)
		} else {
			gpusNotMadeExclusive = append(gpusNotMadeExclusive, gpuId)
		}
	}

	err = gpuTracker.writeGPUTrackerFile(gpusTrackerData)
	if err != nil {
		logger.Log.Printf("Failed to make GPUs exclusive, Error: %v", err)
		fmt.Printf("Failed to make GPUs exclusive, Error: %v\n", err)
		return err
	}

	if len(gpusMadeExclusive) > 0 {
		logger.Log.Printf("GPUs %v have been made exclusive", gpusMadeExclusive)
		fmt.Printf("GPUs %v have been made exclusive\n", gpusMadeExclusive)
	}
	if len(gpusNotMadeExclusive) > 0 {
		logger.Log.Printf("GPUs %v have not been made exclusive because more than one container is currently using it", gpusNotMadeExclusive)
		fmt.Printf("GPUs %v have not been made exclusive because more than one container is currently using it\n", gpusNotMadeExclusive)
	}
	if len(invalidGPUsRange) > 0 {
		logger.Log.Printf("Ignoring %v GPUs Ranges as they are invalid", invalidGPUsRange)
		fmt.Printf("Ignoring %v GPUs Ranges as they are invalid\n", invalidGPUsRange)
	}
	if len(invalidGPUs) > 0 {
		logger.Log.Printf("Ignoring %v GPUs as they are invalid", invalidGPUs)
		fmt.Printf("Ignoring %v GPUs as they are invalid\n", invalidGPUs)
	}

	return nil
}

func (gpuTracker *gpu_tracker_t) MakeGPUsShared(gpus string) (err error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("MakeGPUsShared lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Printf("Recovered in MakeGPUsShared: %v", r)
			err = fmt.Errorf("Panic in MakeGPUsShared: %v", r)
		}
	}()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		logger.Log.Printf("Failed to check if GPU Tracker is initialized, Error:%v", err)
		fmt.Printf("Failed to check if GPU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if !gpuTrackerInitialized {
		err = gpuTracker.initializeGPUTracker()
		if err != nil {
			return err
		}
	}

	gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		logger.Log.Printf("Failed to make GPUs %v shared, Error: %v", gpus, err)
		fmt.Printf("Failed to make GPUs %v shared, Error: %v\n", gpus, err)
		return err
	}

	if gpusTrackerData.Enabled == false {
		logger.Log.Printf("GPU Tracker is disabled")
		fmt.Printf("GPU Tracker is disabled\n")
		return nil
	}

	result, err := gpuTracker.validateGPUsInfo(gpusTrackerData.GPUsInfo)
	if err != nil || result != true {
		return err
	}

	validGPUs, invalidGPUs, invalidGPUsRange, err := gpuTracker.parseGPUsList(gpus)
	if err != nil {
		logger.Log.Printf("Failed to parse GPUs list %v, Error: %v", gpus, err)
		fmt.Printf("Failed to parse GPUs list %v, Error: %v\n", gpus, err)
		return err
	}

	for _, gpuId := range validGPUs {
		gpusTrackerData.GPUsStatus[gpuId] = gpu_status_t{
			UUID:          gpusTrackerData.GPUsStatus[gpuId].UUID,
			PartitionType: gpusTrackerData.GPUsStatus[gpuId].PartitionType,
			Accessibility: SHARED_ACCESS,
			ContainerIds:  gpusTrackerData.GPUsStatus[gpuId].ContainerIds,
		}
	}

	err = gpuTracker.writeGPUTrackerFile(gpusTrackerData)
	if err != nil {
		logger.Log.Printf("Failed to make GPUs shared, Error: %v", err)
		fmt.Printf("Failed to make GPUs shared, Error: %v\n", err)
		return err
	}

	if len(validGPUs) > 0 {
		logger.Log.Printf("GPUs %v have been made shared", validGPUs)
		fmt.Printf("GPUs %v have been made shared\n", validGPUs)
	}
	if len(invalidGPUsRange) > 0 {
		logger.Log.Printf("Ignoring %v GPUs Ranges as they are invalid", invalidGPUsRange)
		fmt.Printf("Ignoring %v GPUs Ranges as they are invalid\n", invalidGPUsRange)
	}
	if len(invalidGPUs) > 0 {
		logger.Log.Printf("Ignoring %v GPUs as they are invalid", invalidGPUs)
		fmt.Printf("Ignoring %v GPUs as they are invalid\n", invalidGPUs)
	}

	return nil
}

func (gpuTracker *gpu_tracker_t) ReserveGPUs(gpus string, containerId string) (allocatedGPUs []int, err error) {
	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile)
	if err != nil {
		return nil, fmt.Errorf("ReserveGPUs lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Printf("Recovered in ReserveGPUs: %v", r)
			err = fmt.Errorf("Panic in ReserveGPUs: %v", r)
			allocatedGPUs = []int{}
		}
	}()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		logger.Log.Printf("Failed to check if GPU Tracker is initialized, Error:%v", err)
		fmt.Printf("Failed to check if GPU Tracker is initialized, Error:%v\n", err)
		return []int{}, err
	}

	if !gpuTrackerInitialized {
		err = gpuTracker.initializeGPUTracker()
		if err != nil {
			return []int{}, err
		}
	}

	gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
	if err != nil {
		logger.Log.Printf("Failed to reserve GPUs %v, Error: %v", gpus, err)
		fmt.Printf("Failed to reserve GPUs %v, Error: %v\n", gpus, err)
		return []int{}, err
	}

	validGPUs, invalidGPUs, invalidGPUsRange, err := gpuTracker.parseGPUsList(gpus)
	if err != nil {
		logger.Log.Printf("Failed to parse GPUs list %v, Error: %v", gpus, err)
		fmt.Printf("Failed to parse GPUs list %v, Error: %v\n", gpus, err)
		return []int{}, err
	}
	if len(invalidGPUsRange) > 0 {
		logger.Log.Printf("Ignoring %v GPUs Ranges as they are invalid", invalidGPUsRange)
		fmt.Printf("Ignoring %v GPUs Ranges as they are invalid\n", invalidGPUsRange)
	}
	if len(invalidGPUs) > 0 {
		logger.Log.Printf("Ignoring %v GPUs as they are invalid", invalidGPUs)
		fmt.Printf("Ignoring %v GPUs as they are invalid\n", invalidGPUs)
	}

	if gpusTrackerData.Enabled == false {
		logger.Log.Printf("GPU Tracker is disabled")
		return validGPUs, nil
	}

	result, err := gpuTracker.validateGPUsInfo(gpusTrackerData.GPUsInfo)
	if err != nil || result != true {
		return []int{}, fmt.Errorf("GPUs info is invalid. Please reset GPU Tracker.\n")
	}

	var unavailableGPUs []int
	for _, gpuId := range validGPUs {
		if gpusTrackerData.GPUsStatus[gpuId].Accessibility == SHARED_ACCESS ||
			(gpusTrackerData.GPUsStatus[gpuId].Accessibility == EXCLUSIVE_ACCESS &&
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

	err = gpuTracker.writeGPUTrackerFile(gpusTrackerData)
	if err != nil {
		logger.Log.Printf("Failed to reserve GPUs %v, Error: %v", validGPUs, err)
		fmt.Printf("Failed to reserve GPUs %v, Error: %v\n", validGPUs, err)
		return []int{}, err
	}

	if len(allocatedGPUs) > 0 {
		logger.Log.Printf("GPUs %v allocated", allocatedGPUs)
		fmt.Printf("GPUs %v allocated\n", allocatedGPUs)
	}
	if len(unavailableGPUs) > 0 {
		logger.Log.Printf("GPUs %v are exlusive and already in use", unavailableGPUs)
		fmt.Printf("GPUs %v are exclusive and already in use\n", unavailableGPUs)
		return []int{}, fmt.Errorf("GPUs %v are exclusive and already in use\n", unavailableGPUs)
	}

	return allocatedGPUs, nil
}

func (gpuTracker *gpu_tracker_t) ReleaseGPUs(containerId string) (err error) {
	removeContainerId := func(containerId string, containerIds []string) ([]string, bool) {
		for idx, id := range containerIds {
			if id == containerId {
				return append(containerIds[:idx], containerIds[idx+1:]...), true
			}
		}
		return containerIds, false
	}

	lock, err := acquireLock(gpuTracker.gpuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("ReleaseGPUs lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Printf("Recovered in ReleaseGPUs: %v", r)
			err = fmt.Errorf("Panic in ReleaseGPUs: %v", r)
		}
	}()

	gpuTrackerInitialized, err := gpuTracker.isGPUTrackerInitialized()
	if err != nil {
		logger.Log.Printf("Failed to check if GPU Tracker is initialized, Error:%v", err)
		fmt.Printf("Failed to check if GPU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if gpuTrackerInitialized {
		gpusTrackerData, err := gpuTracker.readGPUTrackerFile()
		if err != nil {
			logger.Log.Printf("Failed to release GPUs used by container %v, Error: %v", containerId, err)
			fmt.Printf("Failed to release GPUs used by container %v, Error: %v\n", containerId, err)
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

		err = gpuTracker.writeGPUTrackerFile(gpusTrackerData)
		if err != nil {
			logger.Log.Printf("Failed to release GPUs used by container %v, Error: %v", containerId, err)
			fmt.Printf("Failed to release GPUs used by container %v, Error: %v\n", containerId, err)
			return err
		}

		logger.Log.Printf("Released GPUs %v used by container %v", releasedGPUs, containerId)
		fmt.Printf("Released GPUs %v used by container %v\n", releasedGPUs, containerId)
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
