package cdi

import (
	"fmt"
	"testing"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/logger"
	"tags.cncf.io/container-device-interface/specs-go"
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

func setup(t *testing.T) {
	logger.Init(true)
}

func TestInterface(t *testing.T) {
	setup(t)

	spec := specs.Spec{
		Version: "0.6.0",
		Kind:    "amd.com/gpu",
		Devices: []specs.Device{},
	}
	cdi := &cdi_t{
		spec:    spec,
		getGPUs: mockGetAMDGPUs,
		getGPU:  mockGetAMDGPU,
	}

	err := cdi.GenerateSpec()
	Assert(t, err == nil, fmt.Sprintf("failed to generate cdi spec, Err: %v", err))

	cdi.specPath = "/tmp/"
	err = cdi.WriteSpec()
	Assert(t, err == nil, fmt.Sprintf("WriteSpec() returned error %v", err))

	_, err = cdi.ValidateSpec()
	Assert(t, err == nil, fmt.Sprintf("ValidateSpec() returned error %v", err))

	cdi.PrintSpec()
}

func Assert(t *testing.T, b bool, errString string) {
	if !b {
		t.Errorf(errString)
	}
}
