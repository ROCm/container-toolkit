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

package cdi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/logger"
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

	// PrintSpec prints the generated CDI spec on the console
	PrintSpec() error

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
		return &specs.Spec{}, err
	}
	defer file.Close()

	var spec specs.Spec
	err = json.NewDecoder(file).Decode(&spec)
	if err != nil {
		return &specs.Spec{}, err
	}

	return &spec, nil
}

func (cdi *cdi_t) GenerateSpec() error {
	gpus, err := cdi.getGPUs()
	if err != nil {
		logger.Log.Printf("Failed to get GPUs, Err: %v", err)
		return err
	}

	getCDIDevNode := func(gpu string) (specs.DeviceNode, error) {
		d, err := cdi.getGPU(gpu)
		if err != nil {
			logger.Log.Printf("Failed to get details of %v GPU, Err: %v", gpu, err)
			return specs.DeviceNode{}, err
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
	file, err := os.Create(cdi.specPath)
	if err != nil {
		logger.Log.Printf("Error creating file, Error: %v", err)
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cdi.spec); err != nil {
		fmt.Printf("Error encoding JSON: %s\n", err)
		return err
	}

	logger.Log.Printf("Wrote spec to %v", cdi.specPath)
	fmt.Printf("Generated CDI spec: %v\n", cdi.specPath)
	return nil
}

func (cdi *cdi_t) PrintSpec() error {
	prettyJSON, err := json.MarshalIndent(cdi.spec, "", "  ")
	if err != nil {
		logger.Log.Printf("Failed to marshal JSON, Error: %v", err)
		return err
	}

	fmt.Printf(string(prettyJSON))
	fmt.Printf("\n")

	return nil
}

func (cdi *cdi_t) ValidateSpec() (bool, error) {
	fmt.Printf("Validating CDI spec: %v\n", cdi.specPath)

	savedCDISpec, err := readSpecFromFile(cdi.specPath)
	if err != nil {
		fmt.Printf("Failed to parse %v, Err: %v\n", cdi.specPath, err)
		return false, err
	}

	err = cdi.GenerateSpec()
	if err != nil {
		fmt.Printf("Failed to generate current CDI spec, Err: %v", err)
		return false, err
	}

	equal := reflect.DeepEqual(*savedCDISpec, cdi.spec)
	if equal != true {
		logger.Log.Printf("CDI spec: %v is invalid. Please regenerate CDI spec", cdi.specPath)
		fmt.Printf("CDI spec is invalid\nPlease regenerate CDI spec\n")
		return false, nil
	}

	return true, nil
}

func New(sp string) (Interface, error) {
	if sp == "" {
		sp = CDI_SPEC_PATH
	}

	dir := filepath.Dir(sp)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			logger.Log.Printf("Failed to create %v, Err: %v", dir, err)
			return nil, err
		}
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
