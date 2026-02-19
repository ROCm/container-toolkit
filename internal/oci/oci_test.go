package oci

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/opencontainers/runtime-spec/specs-go"
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

func mockReserveGPUs(gpus string, containerId string) ([]int, error) {
	parseGPUsList := func(gpus string) ([]int, []string, []string, error) {
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

	validGPUs, _, _, err := parseGPUsList(gpus)
	return validGPUs, err
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
		reserveGPUs:                 mockReserveGPUs,
	}
	err := oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))

	oci.getAMDEnv()
	expectedDevs := []int{0, 1}
	Assert(t, slices.Equal(oci.amdDevices, expectedDevs), fmt.Sprintf("expected amdDevices %v, got %v", expectedDevs, oci.amdDevices))
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

	err = oci.addGPUDevice(gpu, nil)
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

func TestGetGPUDeviceModeOverride(t *testing.T) {
	tests := []struct {
		env    []string
		want   os.FileMode
		wantOk bool
	}{
		{nil, 0, false},
		{[]string{}, 0, false},
		{[]string{"AMD_GPU_DEVICE_MODE=0666"}, 0o666, true},
		{[]string{"PATH=/bin", "AMD_GPU_DEVICE_MODE=0660"}, 0o660, true},
		{[]string{"AMD_GPU_DEVICE_MODE=0o777"}, 0o777, true},
		{[]string{"AMD_GPU_DEVICE_MODE=0O666"}, 0o666, true}, // capital O
		{[]string{"AMD_GPU_DEVICE_MODE=07777"}, 0, false},   // above 0777, rejected
		{[]string{"AMD_GPU_DEVICE_MODE=01000"}, 0, false},   // above 0777, rejected
		{[]string{"AMD_GPU_DEVICE_MODE=invalid"}, 0, false},
		{[]string{"AMD_GPU_DEVICE_MODE="}, 0, false},
		{[]string{"OTHER_VAR=1"}, 0, false}, // env var not set
	}
	setup(t)
	oci := &oci_t{spec: &specs.Spec{}}
	for _, tt := range tests {
		oci.spec.Process = &specs.Process{Env: tt.env}
		got, ok := oci.getGPUDeviceModeOverride(tt.env)
		if ok != tt.wantOk || (ok && got != tt.want) {
			t.Errorf("getGPUDeviceModeOverride(%v) = (%#o, %v), want (%#o, %v)", tt.env, got, ok, tt.want, tt.wantOk)
		}
	}
}

func TestAddGPUDeviceWithModeOverride(t *testing.T) {
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
		Path:     "/dev/dri/renderD128",
		Major:    226,
		Minor:    128,
		FileMode: 0o660,
		Gid:      44,
		Uid:      0,
		Allow:    true,
		DevType:  "c",
		Access:   "rwm",
	}
	modeOverride := os.FileMode(0o666)
	err = oci.addGPUDevice(gpu, &modeOverride)
	Assert(t, err == nil, fmt.Sprintf("addGPUDevice returned error %v", err))

	// addGPUDevice appends; find the last device with this path (the one we added with override)
	var dev *specs.LinuxDevice
	for i := range oci.spec.Linux.Devices {
		if oci.spec.Linux.Devices[i].Path == gpu.Path {
			dev = &oci.spec.Linux.Devices[i]
		}
	}
	Assert(t, dev != nil, "device not found in spec")
	Assert(t, dev.FileMode != nil, "FileMode is nil")
	Assert(t, *dev.FileMode == 0o666, fmt.Sprintf("expected FileMode 0666, got %#o", *dev.FileMode))
}

// TestAddGPUDevicesWithModeOverrideEnv verifies the full path: AMD_GPU_DEVICE_MODE in spec env
// is parsed in addGPUDevices and applied to all GPU devices added to the spec.
func TestAddGPUDevicesWithModeOverrideEnv(t *testing.T) {
	setup(t)
	oci := &oci_t{
		origSpecPath:                TEST_OCI_SPEC_PATH,
		getGPUs:                     mockGetAMDGPUs,
		getGPU:                      mockGetAMDGPU,
		getUniqueIdToDeviceIndexMap: mockGetUniqueIdToDeviceIndexMap,
		reserveGPUs:                 mockReserveGPUs,
	}
	err := oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))
	oci.spec.Process.Env = append(oci.spec.Process.Env, "AMD_GPU_DEVICE_MODE=0666")

	err = oci.getAMDEnv()
	Assert(t, err == nil, fmt.Sprintf("getAMDEnv returned error %v", err))
	err = oci.addGPUDevices()
	Assert(t, err == nil, fmt.Sprintf("addGPUDevices returned error %v", err))

	// At least one GPU device in the spec should have FileMode 0666 (from env override)
	found := false
	for i := range oci.spec.Linux.Devices {
		d := &oci.spec.Linux.Devices[i]
		if (strings.Contains(d.Path, "/dev/dri") || d.Path == "/dev/kfd") && d.FileMode != nil && *d.FileMode == 0o666 {
			found = true
			break
		}
	}
	Assert(t, found, "expected at least one GPU device with FileMode 0666 from AMD_GPU_DEVICE_MODE env")
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
		reserveGPUs:                 mockReserveGPUs,
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
		reserveGPUs:                 mockReserveGPUs,
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
		reserveGPUs:                 mockReserveGPUs,
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
		reserveGPUs:                 mockReserveGPUs,
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
		reserveGPUs:                 mockReserveGPUs,
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
		reserveGPUs:                 mockReserveGPUs,
	}
	err = oci.getSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to get OCI spec, Err: %v", err))

	oci.getAMDEnv()
	expectedDevs := []int{0, 0} // Both device index 0 and UUID that maps to 0 should result in [0, 0] - duplicates allowed
	Assert(t, slices.Equal(oci.amdDevices, expectedDevs), fmt.Sprintf("expected amdDevices %v, got %v", expectedDevs, oci.amdDevices))
}
