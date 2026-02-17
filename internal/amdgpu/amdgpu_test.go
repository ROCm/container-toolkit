package amdgpu

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	// Initialize logger for tests
	logger.Log = log.New(os.Stderr, "", log.LstdFlags)
}

// Mock filesystem operations
type mockFS struct {
	mock.Mock
}

func (m *mockFS) Stat(name string) (os.FileInfo, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(os.FileInfo), args.Error(1)
}

func (m *mockFS) Glob(pattern string) ([]string, error) {
	args := m.Called(pattern)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockFS) ReadFile(name string) ([]byte, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockFS) GetDeviceStat(dev string, format string) (string, error) {
	args := m.Called(dev, format)
	return args.String(0), args.Error(1)
}

// Mock file info
type mockFileInfo struct {
	mock.Mock
}

func (m *mockFileInfo) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockFileInfo) Size() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *mockFileInfo) Mode() os.FileMode {
	args := m.Called()
	return args.Get(0).(os.FileMode)
}

func (m *mockFileInfo) ModTime() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *mockFileInfo) IsDir() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockFileInfo) Sys() interface{} {
	args := m.Called()
	return args.Get(0)
}

func setupMockFileInfo() *mockFileInfo {
	fileInfo := &mockFileInfo{}
	fileInfo.On("Name").Return("")
	fileInfo.On("Size").Return(int64(0))
	fileInfo.On("Mode").Return(os.FileMode(0))
	fileInfo.On("ModTime").Return(time.Now())
	fileInfo.On("IsDir").Return(true)
	fileInfo.On("Sys").Return(nil)
	return fileInfo
}

func setupTopologyData(t *testing.T, mockFS *mockFS, testCase string) {
	topoDir := filepath.Join("../../tests", "amdgpu", "topology", "nodes")
	var topologyFiles []string

	switch testCase {
	case "single_gpu":
		// Single GPU case - only one node
		topologyFiles = []string{
			filepath.Join(topoDir, "0", "properties"),
		}
	case "gpu_with_partition":
		// GPU with partition case - two nodes
		topologyFiles = []string{
			filepath.Join(topoDir, "0", "properties"),
			filepath.Join(topoDir, "1", "properties"),
		}
	case "multiple_gpus":
		// Multiple GPUs case - two nodes
		topologyFiles = []string{
			filepath.Join(topoDir, "0", "properties"),
			filepath.Join(topoDir, "2", "properties"),
		}
	case "unordered_partitions":
		// Unordered partitions case - three nodes
		topologyFiles = []string{
			filepath.Join(topoDir, "0", "properties"),
			filepath.Join(topoDir, "1", "properties"),
			filepath.Join(topoDir, "2", "properties"),
		}
	}

	mockFS.On("Glob", "/sys/class/kfd/kfd/topology/nodes/*/properties").Return(topologyFiles, nil)
	for _, file := range topologyFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read test data: %v", err)
		}
		mockFS.On("ReadFile", file).Return(content, nil)
	}
}

func loadTestData(t *testing.T, mockFS *mockFS, testCase string) {
	// Load topology data based on test case
	setupTopologyData(t, mockFS, testCase)

	// Setup PCI devices based on test case
	switch testCase {
	case "single_gpu":
		mockFS.On("Glob", "/sys/module/amdgpu/drivers/pci:amdgpu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*").
			Return([]string{"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0"}, nil)
		mockFS.On("Glob", "/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/*").
			Return([]string{
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/card0",
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/renderD128",
			}, nil)
		mockFS.On("Glob", "/sys/devices/platform/amdgpu_xcp_*").Return([]string{}, nil)

	case "gpu_with_partition":
		mockFS.On("Glob", "/sys/module/amdgpu/drivers/pci:amdgpu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*").
			Return([]string{"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0"}, nil)
		mockFS.On("Glob", "/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/*").
			Return([]string{
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/card0",
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/renderD128",
			}, nil)
		mockFS.On("Glob", "/sys/devices/platform/amdgpu_xcp_*").
			Return([]string{"/sys/devices/platform/amdgpu_xcp_0"}, nil)
		mockFS.On("Glob", "/sys/devices/platform/amdgpu_xcp_0/drm/*").
			Return([]string{
				"/sys/devices/platform/amdgpu_xcp_0/drm/card1",
				"/sys/devices/platform/amdgpu_xcp_0/drm/renderD129",
			}, nil)

	case "multiple_gpus":
		mockFS.On("Glob", "/sys/module/amdgpu/drivers/pci:amdgpu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*").
			Return([]string{
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0",
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:02:00.0",
			}, nil)
		mockFS.On("Glob", "/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/*").
			Return([]string{
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/card0",
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/renderD128",
			}, nil)
		mockFS.On("Glob", "/sys/module/amdgpu/drivers/pci:amdgpu/0000:02:00.0/drm/*").
			Return([]string{
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:02:00.0/drm/card1",
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:02:00.0/drm/renderD130",
			}, nil)
		mockFS.On("Glob", "/sys/devices/platform/amdgpu_xcp_*").Return([]string{}, nil)

	case "unordered_partitions":
		mockFS.On("Glob", "/sys/module/amdgpu/drivers/pci:amdgpu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*").
			Return([]string{"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0"}, nil)
		mockFS.On("Glob", "/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/*").
			Return([]string{
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/card0",
				"/sys/module/amdgpu/drivers/pci:amdgpu/0000:01:00.0/drm/renderD128",
			}, nil)
		mockFS.On("Glob", "/sys/devices/platform/amdgpu_xcp_*").
			Return([]string{
				"/sys/devices/platform/amdgpu_xcp_0",
				"/sys/devices/platform/amdgpu_xcp_1",
			}, nil)
		mockFS.On("Glob", "/sys/devices/platform/amdgpu_xcp_0/drm/*").
			Return([]string{
				"/sys/devices/platform/amdgpu_xcp_0/drm/card1",
				"/sys/devices/platform/amdgpu_xcp_0/drm/renderD129",
			}, nil)
		mockFS.On("Glob", "/sys/devices/platform/amdgpu_xcp_1/drm/*").
			Return([]string{
				"/sys/devices/platform/amdgpu_xcp_1/drm/card2",
				"/sys/devices/platform/amdgpu_xcp_1/drm/renderD130",
			}, nil)
	}
}

func TestGetAMDGPUs(t *testing.T) {
	tests := []struct {
		name          string
		testCase      string
		expectedDevs  []DeviceInfo
		expectedError error
	}{
		{
			name:          "no amdgpu driver",
			testCase:      "none",
			expectedDevs:  nil,
			expectedError: os.ErrNotExist,
		},
		{
			name:     "single GPU device",
			testCase: "single_gpu",
			expectedDevs: []DeviceInfo{
				DeviceInfo{
					DrmDevices: []string{
						"/dev/dri/card0",
						"/dev/dri/renderD128",
					},
					PartitionType: "",
				},
			},
			expectedError: nil,
		},
		{
			name:     "GPU with partition",
			testCase: "gpu_with_partition",
			expectedDevs: []DeviceInfo{
				DeviceInfo{
					DrmDevices: []string{
						"/dev/dri/card0",
						"/dev/dri/renderD128",
					},
					PartitionType: "",
				},
				DeviceInfo{
					DrmDevices: []string{
						"/dev/dri/card1",
						"/dev/dri/renderD129",
					},
					PartitionType: "",
				},
			},
			expectedError: nil,
		},
		{
			name:     "multiple GPUs",
			testCase: "multiple_gpus",
			expectedDevs: []DeviceInfo{
				DeviceInfo{
					DrmDevices: []string{
						"/dev/dri/card0",
						"/dev/dri/renderD128",
					},
					PartitionType: "",
				},
				DeviceInfo{
					DrmDevices: []string{
						"/dev/dri/card1",
						"/dev/dri/renderD130",
					},
					PartitionType: "",
				},
			},
			expectedError: nil,
		},
		{
			name:     "unordered partitions",
			testCase: "unordered_partitions",
			expectedDevs: []DeviceInfo{
				DeviceInfo{
					DrmDevices: []string{
						"/dev/dri/card0",
						"/dev/dri/renderD128",
					},
					PartitionType: "",
				},
				DeviceInfo{
					DrmDevices: []string{
						"/dev/dri/card1",
						"/dev/dri/renderD129",
					},
					PartitionType: "",
				},
				DeviceInfo{
					DrmDevices: []string{
						"/dev/dri/card2",
						"/dev/dri/renderD130",
					},
					PartitionType: "",
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &mockFS{}

			if tt.testCase == "none" {
				fileInfo := setupMockFileInfo()
				fileInfo.On("IsDir").Return(false)
				mockFS.On("Stat", "/sys/module/amdgpu/drivers/").Return(nil, os.ErrNotExist)
			} else {
				fileInfo := setupMockFileInfo()
				mockFS.On("Stat", "/sys/module/amdgpu/drivers/").Return(fileInfo, nil)
				loadTestData(t, mockFS, tt.testCase)
			}

			devs, err := GetAMDGPUsWithFS(mockFS)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDevs, devs)
			}

			mockFS.AssertExpectations(t)
		})
	}
}

func TestParseTopologyProperties(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		regex         *regexp.Regexp
		expectedValue int64
		expectedError error
	}{
		{
			name:          "valid property",
			content:       "drm_render_minor 128\n",
			regex:         regexp.MustCompile(`drm_render_minor\s(\d+)`),
			expectedValue: 128,
			expectedError: nil,
		},
		{
			name:          "property not found",
			content:       "some_other_property 123\n",
			regex:         regexp.MustCompile(`drm_render_minor\s(\d+)`),
			expectedValue: 0,
			expectedError: fmt.Errorf("property not found in test.properties"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &mockFS{}
			mockFS.On("ReadFile", "test.properties").Return([]byte(tt.content), nil)

			value, err := ParseTopologyProperties(mockFS, "test.properties", tt.regex)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestParseTopologyPropertiesString(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		regex         *regexp.Regexp
		expectedValue string
		expectedError error
	}{
		{
			name:          "valid property",
			content:       "unique_id 123\n",
			regex:         regexp.MustCompile(`unique_id\s(\w+)`),
			expectedValue: "123",
			expectedError: nil,
		},
		{
			name:          "property not found",
			content:       "some_other_property value\n",
			regex:         regexp.MustCompile(`gpu_id\s(\w+)`),
			expectedValue: "",
			expectedError: fmt.Errorf("property not found in test.properties"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &mockFS{}
			mockFS.On("ReadFile", "test.properties").Return([]byte(tt.content), nil)

			value, err := ParseTopologyPropertiesString(mockFS, "test.properties", tt.regex)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestGetAMDGPUWithFS(t *testing.T) {
	tests := []struct {
		name          string
		device        string
		setupMocks    func(*mockFS)
		expectedGPU   AMDGPU
		expectedError error
	}{
		{
			name:   "valid GPU device",
			device: "/dev/dri/card0",
			setupMocks: func(m *mockFS) {
				m.On("GetDeviceStat", "/dev/dri/card0", "%t").Return("e2", nil)  // major number (226 in hex)
				m.On("GetDeviceStat", "/dev/dri/card0", "%T").Return("0", nil)   // minor number
				m.On("GetDeviceStat", "/dev/dri/card0", "%a").Return("666", nil) // file mode (octal)
				m.On("GetDeviceStat", "/dev/dri/card0", "%g").Return("44", nil)  // group ID
			},
			expectedGPU: AMDGPU{
				Path:     "/dev/dri/card0",
				Major:    226,
				Minor:    0,
				FileMode: 0666,
				Gid:      44,
				Uid:      0,
				Allow:    true,
				DevType:  "c",
				Access:   "rwm",
			},
			expectedError: nil,
		},
		{
			name:   "valid render device",
			device: "/dev/dri/renderD128",
			setupMocks: func(m *mockFS) {
				m.On("GetDeviceStat", "/dev/dri/renderD128", "%t").Return("e2", nil)  // major number (226 in hex)
				m.On("GetDeviceStat", "/dev/dri/renderD128", "%T").Return("80", nil)  // minor number (128 in hex)
				m.On("GetDeviceStat", "/dev/dri/renderD128", "%a").Return("666", nil) // file mode (octal)
				m.On("GetDeviceStat", "/dev/dri/renderD128", "%g").Return("44", nil)  // group ID
			},
			expectedGPU: AMDGPU{
				Path:     "/dev/dri/renderD128",
				Major:    226,
				Minor:    128,
				FileMode: 0666,
				Gid:      44,
				Uid:      0,
				Allow:    true,
				DevType:  "c",
				Access:   "rwm",
			},
			expectedError: nil,
		},
		{
			name:   "non-existent device",
			device: "/dev/dri/card999",
			setupMocks: func(m *mockFS) {
				m.On("GetDeviceStat", "/dev/dri/card999", "%t").Return("", fmt.Errorf("stat failed"))
				m.On("GetDeviceStat", "/dev/dri/card999", "%T").Return("", fmt.Errorf("stat failed"))
				m.On("GetDeviceStat", "/dev/dri/card999", "%a").Return("", fmt.Errorf("stat failed"))
				m.On("GetDeviceStat", "/dev/dri/card999", "%g").Return("", fmt.Errorf("stat failed"))
			},
			expectedGPU: AMDGPU{
				Path:     "/dev/dri/card999",
				Major:    0,
				Minor:    0,
				FileMode: 0,
				Gid:      0,
				Uid:      0,
				Allow:    true,
				DevType:  "c",
				Access:   "rwm",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &mockFS{}
			tt.setupMocks(mockFS)

			gpu, err := GetAMDGPUWithFS(mockFS, tt.device)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedGPU, gpu)
			}

			mockFS.AssertExpectations(t)
		})
	}
}

func TestGetDevIdsFromTopology(t *testing.T) {
	tests := []struct {
		name           string
		testCase       string
		expectedResult map[int]string
	}{
		{
			name:     "single GPU topology",
			testCase: "single_gpu",
			expectedResult: map[int]string{
				128: "1",
			},
		},
		{
			name:     "GPU with partition topology",
			testCase: "gpu_with_partition",
			expectedResult: map[int]string{
				128: "1",
				129: "1",
			},
		},
		{
			name:     "multiple GPUs topology",
			testCase: "multiple_gpus",
			expectedResult: map[int]string{
				128: "1",
				130: "2",
			},
		},
		{
			name:     "unordered partitions topology",
			testCase: "unordered_partitions",
			expectedResult: map[int]string{
				128: "1",
				129: "1",
				130: "2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &mockFS{}
			setupTopologyData(t, mockFS, tt.testCase)

			result := GetDevIdsFromTopology(mockFS, "/sys/class/kfd/kfd")
			assert.Equal(t, tt.expectedResult, result)
			mockFS.AssertExpectations(t)
		})
	}
}

func TestGetUniqueIdToDeviceIndexMapWithFS(t *testing.T) {
	tests := []struct {
		name           string
		testCase       string
		expectedResult map[string][]int
		expectedError  error
	}{
		{
			name:     "single GPU UUID mapping",
			testCase: "single_gpu",
			expectedResult: map[string][]int{
				"0x1": {0},
				"1":   {0},
			},
			expectedError: nil,
		},
		{
			name:     "GPU with partition UUID mapping",
			testCase: "gpu_with_partition",
			expectedResult: map[string][]int{
				"0x1": {0, 1},
				"1":   {0, 1},
			},
			expectedError: nil,
		},
		{
			name:     "multiple GPUs UUID mapping",
			testCase: "multiple_gpus",
			expectedResult: map[string][]int{
				"0x1": {0},
				"1":   {0},
				"0x2": {1},
				"2":   {1},
			},
			expectedError: nil,
		},
		{
			name:     "unordered partitions UUID mapping",
			testCase: "unordered_partitions",
			expectedResult: map[string][]int{
				"0x1": {0, 1},
				"1":   {0, 1},
				"0x2": {2},
				"2":   {2},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &mockFS{}

			// Setup filesystem mocks
			fileInfo := setupMockFileInfo()
			mockFS.On("Stat", "/sys/module/amdgpu/drivers/").Return(fileInfo, nil)
			loadTestData(t, mockFS, tt.testCase)

			result, err := GetUniqueIdToDeviceIndexMapWithFS(mockFS)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockFS.AssertExpectations(t)
		})
	}
}
