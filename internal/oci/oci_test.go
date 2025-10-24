package oci

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/logger"
)

// Constants
const (
	// Test OCI json spec
	TEST_OCI_SPEC_PATH = "../../tests"

	// Create container command with "--bundler xyz" option
	//CREATE_ARGS = "amd-container-runtime --root /var/run/docker/runtime-runc/moby --log /run/containerd/io.containerd.runtime.v2.task/moby/f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919/log.json --log-format json --systemd-cgroup create --bundle /run/containerd/io.containerd.runtime.v2.task/moby/f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919 --pid-file /run/containerd/io.containerd.runtime.v2.task/moby/f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919/init.pid f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919"
	CREATE_ARGS = "amd-container-runtime --root /var/run/docker/runtime-runc/moby --log /run/containerd/io.containerd.runtime.v2.task/moby/f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919/log.json --log-format json --systemd-cgroup create --bundle ../../tests --pid-file /run/containerd/io.containerd.runtime.v2.task/moby/f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919/init.pid f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919"

	// Create container command with "--bundler=xyz" option
	BUNDLE_ARGS = "amd-container-runtime --root /var/run/docker/runtime-runc/moby --log /run/containerd/io.containerd.runtime.v2.task/moby/f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919/log.json --log-format json --systemd-cgroup create --bundle=/run/containerd/io.containerd.runtime.v2.task/moby/f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919 --pid-file /run/containerd/io.containerd.runtime.v2.task/moby/f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919/init.pid f936e9ab998d8dd8000f9f61180754ae669ac89aa594d195ec8a5ef16e1a9919"

	// Delete container command
	DELETE_ARGS = "amd-container-runtime --root /var/run/docker/runtime-runc/moby --log /run/containerd/io.containerd.runtime.v2.task/moby/a557d313712ab3255f3bb0eb107173fe41e386a99e6873311107239d43335085/log.json --log-format json delete --force a557d313712ab3255f3bb0eb107173fe41e386a99e6873311107239d43335085], environ: [LANG=C.UTF-8 PATH=/opt/containerd/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin INVOCATION_ID=56c5fa9ca8514fb7a232d094e1d8bed5 JOURNAL_STREAM=8:228353 SYSTEMD_EXEC_PID=3130 LD_LIBRARY_PATH=/opt/containerd/lib: GOMAXPROCS=2 MAX_SHIM_VERSION=2 TTRPC_ADDRESS=/run/containerd/containerd.sock.ttrpc GRPC_ADDRESS=/run/containerd/containerd.sock NAMESPACE=moby"
)

func mockGetAMDGPUs() ([][]string, error) {
	ret := [][]string{
		{
			"/dev/dri/renderD128",
			"/dev/dri/card1",
		},
		{
			"/dev/dri/render129",
			"/dev/dri/card2",
		},
	}

	return ret, nil
}

func mockGetAMDGPU(dev string) (amdgpu.AMDGPU, error) {
	gpu := amdgpu.AMDGPU{
		Path:     dev,
		Major:    226,
		Minor:    1,
		FileMode: 432,
		Gid:      44,
		Uid:      0,
		Allow:    true,
		DevType:  "c",
		Access:   "rwm",
	}

	return gpu, nil
}

func setup(t *testing.T) {
	logger.Init(true)
}

func TestParseArgs(t *testing.T) {
	setup(t)
	oci := &oci_t{}

	// Empty args
	oci.parseArgs()
	Assert(t, len(oci.args) == 0, fmt.Sprintf("non-empty args, %v", oci.args))
	Assert(t, len(oci.amdDevices) == 0, fmt.Sprintf("non-empty amdDevices, %v", oci.amdDevices))
	Assert(t, !oci.isCreate, "isCreate is True")
	Assert(t, oci.hookPath == "", fmt.Sprintf("non-empty hookPath, %v", oci.hookPath))
	Assert(t, oci.origSpecPath == "", fmt.Sprintf("non-empty origSpecPath, %v", oci.origSpecPath))
	Assert(t, oci.updatedSpecPath == "", fmt.Sprintf("non-empty updatedSpecPath, %v", oci.updatedSpecPath))
	Assert(t, oci.spec == nil, fmt.Sprintf("non-nil spec, %v", oci.spec))

	// Bundle arg ("--bundle xyz") with create command
	oci.args = strings.Split(CREATE_ARGS, " ")
	oci.parseArgs()
	Assert(t, len(oci.args) > 0, fmt.Sprintf("empty args, %v", oci.args))
	Assert(t, oci.isCreate, "isCreate is False")
	Assert(t, oci.origSpecPath != "", "empty origSpecPath")
	Assert(t, oci.updatedSpecPath == oci.origSpecPath, fmt.Sprintf("updateSpecPath %v is different from origSpecPath %v", oci.updatedSpecPath, oci.origSpecPath))

	// Bundle arg ("--bundle=xyz") arg with create command
	oci.args = strings.Split(BUNDLE_ARGS, " ")
	oci.parseArgs()
	Assert(t, len(oci.args) > 0, fmt.Sprintf("empty args, %v", oci.args))
	Assert(t, oci.IsCreate(), "isCreate is False")
	Assert(t, !oci.HasHelpOption(), "hasHelpOption is True")
	Assert(t, oci.origSpecPath != "", "empty origSpecPath")
	Assert(t, oci.updatedSpecPath == oci.origSpecPath, fmt.Sprintf("updateSpecPath %v is different from origSpecPath %v", oci.updatedSpecPath, oci.origSpecPath))

	oci = &oci_t{}
}

func TestGetAMDEnv(t *testing.T) {
	setup(t)
	oci := &oci_t{
		origSpecPath:                TEST_OCI_SPEC_PATH,
		getGPUs:                     mockGetAMDGPUs,
		getGPU:                      mockGetAMDGPU,
		getUniqueIdToDeviceIndexMap: mockGetUniqueIdToDeviceIndexMap,
	}
	err := oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))

	oci.getAMDEnv()
	expectedDevs := []int{0, 1}
	Assert(t, slices.Equal(oci.amdDevices, expectedDevs), fmt.Sprintf("expected amdDevices %v, got %v", expectedDevs, oci.amdDevices))
	Assert(t, !oci.isAddAllGPUs(), "isAddAllGPUs() returned True")
	Assert(t, !oci.isAddNoGPUs(), "isAddNoGPUs() returned True")

	err = oci.addGPUDevices()
	Assert(t, err == nil, fmt.Sprintf("addGPUDevices returned error %v", err))

	oci = &oci_t{}
}

func TestAddGPUDevice(t *testing.T) {
	setup(t)
	oci := &oci_t{
		origSpecPath:                TEST_OCI_SPEC_PATH,
		getGPUs:                     mockGetAMDGPUs,
		getGPU:                      mockGetAMDGPU,
		getUniqueIdToDeviceIndexMap: mockGetUniqueIdToDeviceIndexMap,
	}
	err := oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))

	gpu := amdgpu.AMDGPU{
		Path:     "/dev/dri/card1",
		Major:    226,
		Minor:    1,
		FileMode: 432,
		Gid:      44,
		Uid:      0,
		Allow:    true,
		DevType:  "c",
		Access:   "rwm",
	}

	err = oci.addGPUDevice(gpu)
	Assert(t, err == nil, fmt.Sprintf("addGpuDevice returned error %v", err))

	Assert(t, oci.spec.Linux != nil, "oci.spec.Linux is nil")
	devFound := false
	for _, d := range oci.spec.Linux.Devices {
		if d.Path == "/dev/dri/card1" {
			devFound = true
			Assert(t, d.Type == gpu.DevType, fmt.Sprintf("expected type %v, got type %v", gpu.DevType, d.Type))
			Assert(t, d.Major == gpu.Major, fmt.Sprintf("expected major %v, got major %v", gpu.Major, d.Major))
			Assert(t, d.Minor == gpu.Minor, fmt.Sprintf("expected Minor %v, got Minor %v", gpu.Minor, d.Minor))
			break
		}
	}
	Assert(t, devFound, fmt.Sprintf("dev %v not found in spec", gpu.Path))

	Assert(t, oci.spec.Linux.Resources != nil, "oci.spec.Linux.Resources is nil")
	resDevFound := false
	for _, d := range oci.spec.Linux.Resources.Devices {
		if d.Major != nil && d.Minor != nil &&
			*d.Major == gpu.Major && *d.Minor == gpu.Minor {
			resDevFound = true
			Assert(t, d.Type == gpu.DevType, fmt.Sprintf("expected type %v, got type %v", gpu.DevType, d.Type))
			Assert(t, d.Allow == gpu.Allow, fmt.Sprintf("expected allow %v, got allow %v", gpu.Allow, d.Allow))
			break
		}
	}
	Assert(t, resDevFound, fmt.Sprintf("dev %v,%v not found in spec", gpu.Major, gpu.Minor))
}

func TestNew(t *testing.T) {
	_, err := New(strings.Split(CREATE_ARGS, " "))
	Assert(t, err == nil, fmt.Sprintf("New() returned error %v", err))
}
func TestInterface(t *testing.T) {
	setup(t)

	oci := &oci_t{
		origSpecPath:                TEST_OCI_SPEC_PATH,
		getGPUs:                     mockGetAMDGPUs,
		getGPU:                      mockGetAMDGPU,
		getUniqueIdToDeviceIndexMap: mockGetUniqueIdToDeviceIndexMap,
	}
	err := oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))

	err = oci.UpdateSpec(AddGPUDevices)
	Assert(t, err == nil, fmt.Sprintf("UpdateSpec(AddGPUDevices) returned error %v", err))

	err = oci.UpdateSpec(AddHook)
	Assert(t, err == nil, fmt.Sprintf("UpdateSpec(AddHook) returned error %v", err))

	oci.updatedSpecPath = "/tmp"
	err = oci.WriteSpec()
	Assert(t, err == nil, fmt.Sprintf("WriteSpec() returned error %v", err))

	oci.PrintSpec()
}

func Assert(t *testing.T, b bool, errString string) {
	if !b {
		t.Errorf(errString)
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid hex string lowercase",
			input:    "abc123",
			expected: true,
		},
		{
			name:     "valid hex string uppercase",
			input:    "ABC123",
			expected: true,
		},
		{
			name:     "valid hex string mixed case",
			input:    "aBc123DeF",
			expected: true,
		},
		{
			name:     "invalid hex string with g",
			input:    "abc123g",
			expected: false,
		},
		{
			name:     "invalid hex string with special chars",
			input:    "abc123-def",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "single valid hex char",
			input:    "a",
			expected: true,
		},
		{
			name:     "long valid hex string",
			input:    "ef2c1799a1f3e2ed",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHexString(tt.input)
			Assert(t, result == tt.expected, fmt.Sprintf("isHexString(%s) = %t, expected %t", tt.input, result, tt.expected))
		})
	}
}

// Mock for GetUniqueIdToDeviceIndexMap
func mockGetUniqueIdToDeviceIndexMap() (map[string][]int, error) {
	return map[string][]int{
		"0xef2c1799a1f3e2ed": {0},
		"ef2c1799a1f3e2ed":   {0},
		"0x1234567890abcdef": {1},
		"1234567890abcdef":   {1},
		"0xpartitionedgpu":   {0, 1}, // Example partitioned GPU with multiple indices
		"partitionedgpu":     {0, 1},
	}, nil
}

func TestGetAMDEnvWithUUID(t *testing.T) {
	setup(t)

	// Test with hex UUID in AMD_VISIBLE_DEVICES
	testSpec := `{
		"process": {
			"env": [
				"AMD_VISIBLE_DEVICES=0xef2c1799a1f3e2ed,0x1234567890abcdef"
			]
		}
	}`

	// Create a temporary test file
	tmpDir := t.TempDir()
	specPath := tmpDir + "/config.json"
	err := os.WriteFile(specPath, []byte(testSpec), 0644)
	Assert(t, err == nil, fmt.Sprintf("failed to write test spec, Err: %v", err))

	oci := &oci_t{
		origSpecPath:                tmpDir,
		getGPUs:                     mockGetAMDGPUs,
		getGPU:                      mockGetAMDGPU,
		getUniqueIdToDeviceIndexMap: mockGetUniqueIdToDeviceIndexMap,
	}
	err = oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))

	oci.getAMDEnv()
	expectedDevs := []int{0, 1}
	Assert(t, slices.Equal(oci.amdDevices, expectedDevs), fmt.Sprintf("expected amdDevices %v, got %v", expectedDevs, oci.amdDevices))
}

func TestGetAMDEnvWithDockerResource(t *testing.T) {
	setup(t)

	// Test with DOCKER_RESOURCE_GPU containing hex UUIDs
	testSpec := `{
		"process": {
			"env": [
				"DOCKER_RESOURCE_GPU=ef2c1799a1f3e2ed"
			]
		}
	}`

	// Create a temporary test file
	tmpDir := t.TempDir()
	specPath := tmpDir + "/config.json"
	err := os.WriteFile(specPath, []byte(testSpec), 0644)
	Assert(t, err == nil, fmt.Sprintf("failed to write test spec, Err: %v", err))

	oci := &oci_t{
		origSpecPath:                tmpDir,
		getGPUs:                     mockGetAMDGPUs,
		getGPU:                      mockGetAMDGPU,
		getUniqueIdToDeviceIndexMap: mockGetUniqueIdToDeviceIndexMap,
	}
	err = oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))

	oci.getAMDEnv()
	expectedDevs := []int{0}
	Assert(t, slices.Equal(oci.amdDevices, expectedDevs), fmt.Sprintf("expected amdDevices %v, got %v", expectedDevs, oci.amdDevices))
}

func TestGetAMDEnvWithMixedDevices(t *testing.T) {
	setup(t)

	// Test with mixed device indices and UUIDs (different indices)
	testSpec := `{
		"process": {
			"env": [
				"AMD_VISIBLE_DEVICES=1,0xef2c1799a1f3e2ed"
			]
		}
	}`

	// Create a temporary test file
	tmpDir := t.TempDir()
	specPath := tmpDir + "/config.json"
	err := os.WriteFile(specPath, []byte(testSpec), 0644)
	Assert(t, err == nil, fmt.Sprintf("failed to write test spec, Err: %v", err))

	oci := &oci_t{
		origSpecPath:                tmpDir,
		getGPUs:                     mockGetAMDGPUs,
		getGPU:                      mockGetAMDGPU,
		getUniqueIdToDeviceIndexMap: mockGetUniqueIdToDeviceIndexMap,
	}
	err = oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))

	oci.getAMDEnv()
	expectedDevs := []int{0, 1} // UUID maps to 0, device index is 1, sorted result should be [0, 1]
	Assert(t, slices.Equal(oci.amdDevices, expectedDevs), fmt.Sprintf("expected amdDevices %v, got %v", expectedDevs, oci.amdDevices))
}

func TestGetAMDEnvWithInvalidUUID(t *testing.T) {
	setup(t)

	// Test with invalid UUID that doesn't exist in mapping
	testSpec := `{
		"process": {
			"env": [
				"AMD_VISIBLE_DEVICES=0xdeadbeefdeadbeef"
			]
		}
	}`

	// Create a temporary test file
	tmpDir := t.TempDir()
	specPath := tmpDir + "/config.json"
	err := os.WriteFile(specPath, []byte(testSpec), 0644)
	Assert(t, err == nil, fmt.Sprintf("failed to write test spec, Err: %v", err))

	oci := &oci_t{
		origSpecPath:                tmpDir,
		getGPUs:                     mockGetAMDGPUs,
		getGPU:                      mockGetAMDGPU,
		getUniqueIdToDeviceIndexMap: mockGetUniqueIdToDeviceIndexMap,
	}
	err = oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))

	oci.getAMDEnv()
	expectedDevs := []int{} // Invalid UUID should result in no devices
	Assert(t, slices.Equal(oci.amdDevices, expectedDevs), fmt.Sprintf("expected amdDevices %v, got %v", expectedDevs, oci.amdDevices))
}

func TestGetAMDEnvWithDuplicateDevices(t *testing.T) {
	setup(t)

	// Test with duplicate device specification (same device via index and UUID)
	testSpec := `{
		"process": {
			"env": [
				"AMD_VISIBLE_DEVICES=0,0xef2c1799a1f3e2ed"
			]
		}
	}`

	// Create a temporary test file
	tmpDir := t.TempDir()
	specPath := tmpDir + "/config.json"
	err := os.WriteFile(specPath, []byte(testSpec), 0644)
	Assert(t, err == nil, fmt.Sprintf("failed to write test spec, Err: %v", err))

	oci := &oci_t{
		origSpecPath:                tmpDir,
		getGPUs:                     mockGetAMDGPUs,
		getGPU:                      mockGetAMDGPU,
		getUniqueIdToDeviceIndexMap: mockGetUniqueIdToDeviceIndexMap,
	}
	err = oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))

	oci.getAMDEnv()
	expectedDevs := []int{0, 0} // Both device index 0 and UUID that maps to 0 should result in [0, 0] - duplicates allowed
	Assert(t, slices.Equal(oci.amdDevices, expectedDevs), fmt.Sprintf("expected amdDevices %v, got %v", expectedDevs, oci.amdDevices))
}
