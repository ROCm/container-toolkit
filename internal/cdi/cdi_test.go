package cdi

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/stretchr/testify/assert"
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

func TestInterface(t *testing.T) {
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

	cdi.specPath = t.TempDir() + "/amd.json"
	err = cdi.WriteSpec()
	Assert(t, err == nil, fmt.Sprintf("WriteSpec() returned error %v", err))

	_, err = cdi.ValidateSpec()
	Assert(t, err == nil, fmt.Sprintf("ValidateSpec() returned error %v", err))

	_, err = cdi.FormatSpec()
	Assert(t, err == nil, fmt.Sprintf("FormatSpec() returned error %v", err))
}

// dummySpec is a minimal spec used by WriteSpec tests.
var dummySpec = specs.Spec{
	Version: "0.6.0",
	Kind:    "amd.com/gpu",
	Devices: []specs.Device{},
}

func TestWriteSpec(t *testing.T) {

	tests := []struct {
		name    string
		pathRel string
		spec    specs.Spec
		wantErr bool
		setup   func(t *testing.T, dir string)
	}{
		{
			name:    "dir_exists",
			pathRel: "amd.json",
			spec:    dummySpec,
			wantErr: false,
		},
		{
			name:    "creates_nested_dir",
			pathRel: filepath.Join("nested", "deep", "path", "amd.json"),
			spec:    dummySpec,
			wantErr: false,
		},
		{
			name:    "empty_spec",
			pathRel: "amd.json",
			spec:    specs.Spec{},
			wantErr: false,
		},
		{
			name:    "parent_is_file_fails",
			pathRel: filepath.Join("blocker", "spec.json"),
			spec:    dummySpec,
			wantErr: true,
			setup: func(t *testing.T, dir string) {
				if err := os.WriteFile(filepath.Join(dir, "blocker"), []byte("x"), 0644); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, dir)
			}
			specPath := filepath.Join(dir, tt.pathRel)

			cdi := &cdi_t{
				spec:     tt.spec,
				specPath: specPath,
				getGPUs:  mockGetAMDGPUs,
				getGPU:   mockGetAMDGPU,
			}
			err := cdi.WriteSpec()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			_, err = os.Stat(specPath)
			assert.NoError(t, err, "spec file should be created")
			read, err := readSpecFromFile(specPath)
			assert.NoError(t, err)
			assert.Equal(t, tt.spec, *read, "written spec should match original")
		})
	}
}

func TestWriteSpec_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "amd.json")

	cdi := &cdi_t{
		spec:     dummySpec,
		specPath: specPath,
		getGPUs:  mockGetAMDGPUs,
		getGPU:   mockGetAMDGPU,
	}
	assert.NoError(t, cdi.WriteSpec(), "first WriteSpec")

	second := specs.Spec{
		Version: "0.6.0",
		Kind:    "amd.com/gpu",
		Devices: []specs.Device{{Name: "second", ContainerEdits: specs.ContainerEdits{}}},
	}
	cdi.spec = second
	assert.NoError(t, cdi.WriteSpec(), "second WriteSpec")

	read, err := readSpecFromFile(specPath)
	assert.NoError(t, err)
	assert.Equal(t, second, *read, "file should contain second spec after overwrite")
}

func Assert(t *testing.T, b bool, errString string) {
	if !b {
		t.Errorf(errString)
	}
}
