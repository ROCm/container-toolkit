package gpuTracker

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/logger"
)

func mockIsGPUTrackerInitialized() (bool, error) {
	return true, nil
}

func mockInitializeGPUTracker() error {
	return nil
}

// Mock for GetAMDGPUs
func mockGetAMDGPUs() ([]amdgpu.DeviceInfo, error) {
	ret := []amdgpu.DeviceInfo{
		{
			DrmDevices: []string{
				"/dev/dri/renderD128",
				"/dev/dri/card1",
			},
			PartitionType: "",
		},
		{
			DrmDevices: []string{
				"/dev/dri/renderD129",
				"/dev/dri/card2",
			},
			PartitionType: "",
		},
	}

	return ret, nil
}

// Mock for GetUniqueIdToDeviceIndexMap
func mockGetUniqueIdToDeviceIndexMap() (map[string][]int, error) {
	return map[string][]int{
		"0xef2c1799a1f3e2ed": {0},
		"ef2c1799a1f3e2ed":   {0},
		"0x1234567890abcdef": {1},
		"1234567890abcdef":   {1},
	}, nil
}

func mockParseGPUsList(gpus string) ([]int, []string, []string, error) {
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

	gpusInfo, err := mockGetAMDGPUs()
	if err != nil {
		logger.Log.Printf("Failed to get AMD GPUs info, Error: %v", err)
		fmt.Printf("Failed to get AMD GPUs info, Error: %v\n", err)
		return []int{}, []string{}, []string{}, err
	}

	if gpus == "all" || gpus == "All" || gpus == "ALL" {
		for i := 0; i < len(gpusInfo); i++ {
			validGPUs = append(validGPUs, i)
		}
		return validGPUs, []string{}, []string{}, nil
	}

	uuidToGPUIdMap, err := mockGetUniqueIdToDeviceIndexMap()
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

func mockReadGPUTrackerFile() (gpu_tracker_data_t, error) {
	return gpu_tracker_data_t{
		Enabled: true,
		GPUsStatus: map[int]gpu_status_t{
			0: {
				UUID:          "0xef2c1799a1f3e2ed",
				PartitionType: "",
				Accessibility: 0,
				ContainerIds:  []string{"container_1", "container_2"},
			},
			1: {
				UUID:          "0x1234567890abcdef",
				PartitionType: "",
				Accessibility: 1,
				ContainerIds:  []string{"container_1"},
			},
		},
		GPUsInfo: map[int]amdgpu.DeviceInfo{
			0: {
				DrmDevices: []string{
					"/dev/dri/renderD128",
					"/dev/dri/card1",
				},
				PartitionType: "",
			},
			1: {
				DrmDevices: []string{
					"/dev/dri/renderD129",
					"/dev/dri/card2",
				},
				PartitionType: "",
			},
		},
	}, nil
}

func mockWriteGPUTrackerFile(gpu_tracker_data_t) error {
	return nil
}

func mockValidateGPUsInfo(map[int]amdgpu.DeviceInfo) (bool, error) {
	return true, nil
}

func setup(t *testing.T) {
	logger.Init(true)
}

func TestInterface(t *testing.T) {
	setup(t)

	gpuTracker := &gpu_tracker_t{
		gpuTrackerLockFile:      "/tmp/gpu-tracker.lock",
		isGPUTrackerInitialized: mockIsGPUTrackerInitialized,
		initializeGPUTracker:    mockInitializeGPUTracker,
		parseGPUsList:           mockParseGPUsList,
		readGPUTrackerFile:      mockReadGPUTrackerFile,
		writeGPUTrackerFile:     mockWriteGPUTrackerFile,
		validateGPUsInfo:        mockValidateGPUsInfo,
	}

	err := gpuTracker.Init()
	Assert(t, err == nil, fmt.Sprintf("Init() returned error %v", err))

	err = gpuTracker.Enable()
	Assert(t, err == nil, fmt.Sprintf("Enable() returned error %v", err))

	err = gpuTracker.Disable()
	Assert(t, err == nil, fmt.Sprintf("Disable() returned error %v", err))

	err = gpuTracker.ShowStatus()
	Assert(t, err == nil, fmt.Sprintf("ShowStatus() returned error %v", err))

	err = gpuTracker.MakeGPUsExclusive("0,1")
	Assert(t, err == nil, fmt.Sprintf("MakeGPUsExclusive() returned error %v", err))

	err = gpuTracker.MakeGPUsShared("0-1")
	Assert(t, err == nil, fmt.Sprintf("MakeGPUsShared() returned error %v", err))

	// Reserve Shared GPU
	_, err = gpuTracker.ReserveGPUs("0xef2c1799a1f3e2ed", "container_3")
	Assert(t, err == nil, fmt.Sprintf("ReserveGPUs() returned error %v", err))

	// Reserve Exclusive GPU that is already assigned
	_, err = gpuTracker.ReserveGPUs("0x1234567890abcdef", "container_3")
	Assert(t, err != nil, fmt.Sprintf("ReserveGPUs() did not returned error when expected"))

	// Reserve Shared and Exclusive GPU that are already assigned
	_, err = gpuTracker.ReserveGPUs("0,0x1234567890abcdef", "container_3")
	Assert(t, err != nil, fmt.Sprintf("ReserveGPUs() did not returned error when expected"))

	err = gpuTracker.ReleaseGPUs("container_1")
	Assert(t, err == nil, fmt.Sprintf("ReleaseGPUs() returned error %v", err))
}

func Assert(t *testing.T, b bool, errString string) {
	if !b {
		t.Errorf(errString)
	}
}
