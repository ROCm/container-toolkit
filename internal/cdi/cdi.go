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
	"strconv"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/logger"
	"tags.cncf.io/container-device-interface/specs-go"
)

// Constants
const (
	// Default path and name for CDI spec
	CDI_SPEC_PATH = "/var/run/cdi"

	// CDI spec file name
	CDI_SPEC = "amd.json"
)

// GetGPUs is the type for functions that return the lists of all the GPU devices on the system
type GetGPUs func() ([][]string, error)

// GetGPU is the type for functions that return the device information for the given GPU
type GetGPU func(string) (amdgpu.AMDGPU, error)

// Interface for CDI package
type Interface interface {
	// GenerateSpec generates the CDI spec for all GPUs available on the host system
	GenerateSpec() error

	// WriteSpec writes the generated spec to disk
	WriteSpec() error

	// PrintSpec prints the generated CDI spec on the console
	PrintSpec() error
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
		for _, gpu := range gpuList {
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

func (cdi *cdi_t) WriteSpec() error {
	f := cdi.specPath + "/" + CDI_SPEC

	file, err := os.Create(f)
	if err != nil {
		logger.Log.Printf("Error creating file, Error: %v", err)
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(cdi.spec); err != nil {
		fmt.Printf("Error encoding JSON: %s\n", err)
		return err
	}

	logger.Log.Printf("Wrote spec to %v", f)
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

func New(sp string) (Interface, error) {
	if sp == "" {
		sp = CDI_SPEC_PATH
	}

	if _, err := os.Stat(sp); os.IsNotExist(err) {
		err := os.Mkdir(sp, 0755)
		if err != nil {
			logger.Log.Printf("Failed to create %v, Err: %v", sp, err)
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
