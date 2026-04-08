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

package cdi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"tags.cncf.io/container-device-interface/specs-go"
)

// Constants
const (
	// Default full CDI spec path
	CDI_SPEC_PATH = "/var/run/cdi/amd.json"
)

// GetGPUs is the type for functions that return the lists of all the GPU devices on the system
type GetGPUs func() ([]amdgpu.DeviceInfo, error)

// GetGPU is the type for functions that return the device information for the given GPU
type GetGPU func(string) (amdgpu.AMDGPU, error)

// Interface for CDI package
type Interface interface {
	// GenerateSpec generates the CDI spec for all GPUs available on the host system
	GenerateSpec() error

	// GetSpec returns the CDI Spec to the caller
	GetSpec() specs.Spec

	// WriteSpec writes the generated spec to disk
	WriteSpec() error

	// FormatSpec returns the generated CDI spec as a formatted JSON string
	FormatSpec() (string, error)

	// ValidateSpec validated the existing CDI spec on the disk
	ValidateSpec() (bool, error)
}

// cdi_t implements the CDI interface
type cdi_t struct {
	// spec contains the generated CDI spec
	spec specs.Spec

	// specPath is the directory where CDI spec is written to
	specPath string

	// getGPUs is the function that returns the list of GPUs in the system
	getGPUs GetGPUs

	// getGPU is the function that returns the device info of the given GPU
	getGPU GetGPU
}

func readSpecFromFile(f string) (*specs.Spec, error) {
	file, err := os.Open(f)
	if err != nil {
		return &specs.Spec{}, fmt.Errorf("opening CDI spec file %s: %w", f, err)
	}
	defer file.Close()

	var spec specs.Spec
	err = json.NewDecoder(file).Decode(&spec)
	if err != nil {
		return &specs.Spec{}, fmt.Errorf("decoding CDI spec file %s: %w", f, err)
	}

	return &spec, nil
}

func (cdi *cdi_t) GenerateSpec() error {
	gpus, err := cdi.getGPUs()
	if err != nil {
		return fmt.Errorf("getting GPUs: %w", err)
	}

	getCDIDevNode := func(gpu string) (specs.DeviceNode, error) {
		d, err := cdi.getGPU(gpu)
		if err != nil {
			return specs.DeviceNode{}, fmt.Errorf("getting GPU details for %s: %w", gpu, err)
		}

		dn := specs.DeviceNode{
			Path:        d.Path,
			Type:        d.DevType,
			Major:       d.Major,
			Minor:       d.Minor,
			FileMode:    &d.FileMode,
			Permissions: d.Access,
			UID:         &d.Uid,
			GID:         &d.Gid,
		}

		return dn, nil
	}

	kfdDeviceNode, err := getCDIDevNode("/dev/kfd")
	if err != nil {
		return err
	}

	cdiDevs := []specs.Device{}
	for i, gpuList := range gpus {
		devName := strconv.Itoa(i)
		dnl := []*specs.DeviceNode{}
		for _, gpu := range gpuList.DrmDevices {
			dn, err := getCDIDevNode(gpu)
			if err != nil {
				return err
			}
			dnl = append(dnl, &dn)
		}
		dnl = append(dnl, &kfdDeviceNode)
		cdiDev := specs.Device{
			Name: devName,
			ContainerEdits: specs.ContainerEdits{
				DeviceNodes: dnl,
			},
		}
		cdiDevs = append(cdiDevs, cdiDev)
	}

	allDNs := []*specs.DeviceNode{}
	for _, cd := range cdiDevs {
		dnl := cd.ContainerEdits.DeviceNodes
		allDNs = append(allDNs, dnl[:len(dnl)-1]...)
	}
	allDNs = append(allDNs, &kfdDeviceNode)

	allCdiDev := specs.Device{
		Name: "all",
		ContainerEdits: specs.ContainerEdits{
			DeviceNodes: allDNs,
		},
	}
	cdiDevs = append(cdiDevs, allCdiDev)
	cdi.spec.Devices = cdiDevs

	return nil
}

func (cdi *cdi_t) GetSpec() specs.Spec {
	return cdi.spec
}

func (cdi *cdi_t) WriteSpec() error {
	dir := filepath.Dir(cdi.specPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	file, err := os.Create(cdi.specPath)
	if err != nil {
		return fmt.Errorf("creating CDI spec file %s: %w", cdi.specPath, err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cdi.spec); err != nil {
		return fmt.Errorf("encoding CDI spec to %s: %w", cdi.specPath, err)
	}

	return nil
}

func (cdi *cdi_t) FormatSpec() (string, error) {
	prettyJSON, err := json.MarshalIndent(cdi.spec, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling CDI spec to JSON: %w", err)
	}

	return string(prettyJSON), nil
}

func (cdi *cdi_t) ValidateSpec() (bool, error) {
	savedCDISpec, err := readSpecFromFile(cdi.specPath)
	if err != nil {
		return false, err
	}

	err = cdi.GenerateSpec()
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(*savedCDISpec, cdi.spec), nil
}

func New(sp string) (Interface, error) {
	if sp == "" {
		sp = CDI_SPEC_PATH
	}

	spec := specs.Spec{
		Version: "0.6.0",
		Kind:    "amd.com/gpu",
		Devices: []specs.Device{},
	}

	cdi := &cdi_t{
		spec:     spec,
		specPath: sp,
		getGPUs:  amdgpu.GetAMDGPUs,
		getGPU:   amdgpu.GetAMDGPU,
	}

	return cdi, nil
}
