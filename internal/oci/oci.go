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

package oci

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// constants
const (
	// Default path for AMD Container Runtime OCI hook
	DEFAULT_HOOK_PATH = "/usr/bin/amd-container-runtime-hook"
)

// Interface for OCI package
type Interface interface {
	// IsCreate returns true if the container is getting created now
	IsCreate() bool

	// UpdateSpec updates the input OCI spec as per the request op
	UpdateSpec(op SpecUpdateOp) error

	// WriteSpec writes the updated spec back to disk
	WriteSpec() error

	// PrintSpec prints the current spec on the console
	PrintSpec() error
}

// oci_t implements the OCI interface
type oci_t struct {
	// args are the arguments to runtime
	args []string

	// amdDevices lists the AMD devices, requested via the "AMD_VISIBLE_DEVICES
	// ENV variable. The devices are specified by their indices and stored in
	// the ascending order.
	amdDevices []int

	// isCreate specifies if the container is getting created now
	isCreate bool

	// hookPath is the where the OCI hook executable is on the disk
	hookPath string

	// origSpecPath is where the input OCI spec is on the disk
	origSpecPath string

	// updatedSpecPath is where the updated OCI spec is put on the disk
	updatedSpecPath string

	// spec is the structure into which the input spec file is read into
	spec *specs.Spec
}

// SpecUpdateOp specifies type of update operation on the OCI spec
type SpecUpdateOp int

// Enumeration of the update operations on the OCI spec
const (
	AddHook SpecUpdateOp = iota
	AddGPUDevices
)

// parseArgs parses the arguments passed to runtime
func (oci *oci_t) parseArgs() {
	isBundlePathOption := func(arg string) bool {
		if arg == "-b" || arg == "-bundle" || arg == "--b" || arg == "--bundle" {
			return true
		}
		return false
	}

	args := oci.args
	for i := 0; i < len(args)-1; i++ {
		parts := strings.SplitN(args[i], "=", 2)
		if isBundlePathOption(parts[0]) {
			if len(parts) == 2 {
				oci.origSpecPath = parts[1]
			} else {
				oci.origSpecPath = args[i+1]
				i++
			}
		} else if args[i] == "create" {
			oci.isCreate = true
		}
	}

	// By default, updateSpecPath is the same as origSpecPath
	oci.updatedSpecPath = oci.origSpecPath
}

// getAMDEnv reads the value of "AMD_VISIBLE_DEVICES" environment variable
// in the spec.
func (oci *oci_t) getAMDEnv() {
	getDevs := func(devs string) []int {
		dl := []int{}

		if devs == "all" || devs == "All" || devs == "ALL" {
			dl = append(dl, -1)
			return dl
		}

		for _, c := range strings.Split(devs, ",") {
			i, err := strconv.Atoi(c)
			if err == nil && i >= 0 {
				dl = append(dl, i)
			}
		}

		sort.Ints(dl)

		return dl
	}

	if oci.spec != nil && oci.spec.Process != nil {
		envs := oci.spec.Process.Env
		for _, env := range envs {
			pts := strings.SplitN(env, "=", 2)
			if len(pts) == 2 && pts[0] == "AMD_VISIBLE_DEVICES" {
				oci.amdDevices = getDevs(pts[1])
			}
		}
	}
}

// getSpec reads the input OCI spec file into memory
func (oci *oci_t) getSpec() error {
	if len(oci.origSpecPath) == 0 {
		logger.Log.Printf("Spec path is not set")
		return nil
	}

	f := oci.origSpecPath + "/config.json"

	file, err := os.Open(f)
	if err != nil {
		logger.Log.Printf("Error opening file, Error: %v", err)
		return err
	}

	defer file.Close()

	var spec specs.Spec
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&spec); err != nil {
		logger.Log.Printf("Failed to decode JSON, Error: %v", err)
		return err
	}

	oci.spec = &spec

	return nil
}

// addHook adds the AMD runtime OCI hook into the spec
func (oci *oci_t) addHook() error {
	if oci.spec.Hooks == nil {
		oci.spec.Hooks = &specs.Hooks{}
	}

	hook := specs.Hook{
		Path: oci.hookPath,
	}

	oci.spec.Hooks.CreateRuntime = append(oci.spec.Hooks.CreateRuntime, hook)
	logger.Log.Printf("Added OCI runtime hook, %v", oci.hookPath)

	return nil
}

// isAddAllGPUs returns true if all GPUs must be added to OCI spec
func (oci *oci_t) isAddAllGPUs() bool {
	if len(oci.amdDevices) == 1 && oci.amdDevices[0] == -1 {
		return true
	}

	return false
}

// isAddNoGPUs returns true if no GPUs need to be added to OCI spec
func (oci *oci_t) isAddNoGPUs() bool {
	if len(oci.amdDevices) == 0 {
		return true
	}

	return false
}

// addGPUDevices adds requested GPUs to the OCI spec
func (oci *oci_t) addGPUDevices() error {
	addGpus := func(gpus []string) error {
		for _, gpu := range gpus {
			amdGPU, err := amdgpu.GetAMDGPU(gpu)
			if err != nil {
				return err
			}

			err = oci.addGPUDevice(amdGPU)
			if err != nil {
				return err
			}
		}

		return nil
	}

	if oci.isAddNoGPUs() {
		logger.Log.Printf("No GPUs to be added to OCI spec")
		return nil
	}

	devs, err := amdgpu.GetAMDGPUs()
	if err != nil {
		return err
	}

	j, cnt := 0, 0
	for i, gpus := range devs {
		if oci.isAddAllGPUs() {
			addGpus(gpus)
			cnt++
		} else if oci.amdDevices[j] < len(gpus) && j == i {
			addGpus(gpus)
			cnt++
			j++

			if j == len(oci.amdDevices) {
				break
			}
		} else if oci.amdDevices[j] >= len(gpus) {
			break
		}
	}

	if cnt > 0 {
		kfd, err := amdgpu.GetAMDGPU("/dev/kfd")
		if err != nil {
			return err
		}
		oci.addGPUDevice(kfd)
	}

	return nil
}

// addGPUDevice adds the requested GPU device to the OCI spec
func (oci *oci_t) addGPUDevice(gpu amdgpu.AMDGPU) error {
	dev := specs.LinuxDevice{
		Path:     gpu.Path,
		Type:     gpu.DevType,
		Major:    gpu.Major,
		Minor:    gpu.Minor,
		FileMode: &gpu.FileMode,
		GID:      &gpu.Gid,
		UID:      &gpu.Uid,
	}

	if oci.spec.Linux == nil {
		oci.spec.Linux = &specs.Linux{}
	}

	oci.spec.Linux.Devices = append(oci.spec.Linux.Devices, dev)

	rdev := specs.LinuxDeviceCgroup{
		Allow:  gpu.Allow,
		Type:   gpu.DevType,
		Major:  &gpu.Major,
		Minor:  &gpu.Minor,
		Access: gpu.Access,
	}

	if oci.spec.Linux.Resources == nil {
		oci.spec.Linux.Resources = &specs.LinuxResources{}
	}

	oci.spec.Linux.Resources.Devices = append(oci.spec.Linux.Resources.Devices, rdev)
	logger.Log.Printf("Added GPU device %v to OCI spec", gpu.Path)

	return nil
}

// New creates an OCI instance
func New(argv []string) (Interface, error) {
	oci := &oci_t{
		args:     argv,
		hookPath: DEFAULT_HOOK_PATH,
	}

	oci.parseArgs()
	err := oci.getSpec()
	if err != nil {
		return nil, err
	}

	oci.getAMDEnv()

	return oci, nil
}

// IsCreate returns true if the container is getting created now
func (oci *oci_t) IsCreate() bool {
	return oci.isCreate
}

// WriteSpec writes the updated spec back to disk
func (oci *oci_t) WriteSpec() error {
	f := oci.updatedSpecPath + "/config.json"

	file, err := os.Create(f)
	if err != nil {
		logger.Log.Printf("Error creating file, Error: %v", err)
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(oci.spec); err != nil {
		fmt.Printf("Error encoding JSON: %s\n", err)
		return err
	}

	logger.Log.Printf("Wrote spec to %v", f)
	return nil
}

// UpdateSpec updates the input OCI spec as per the request op
func (oci *oci_t) UpdateSpec(op SpecUpdateOp) error {
	switch op {
	case AddHook:
		return oci.addHook()
	case AddGPUDevices:
		return oci.addGPUDevices()
	}

	return nil
}

// PrintSpec prints the current spec on the console
func (oci *oci_t) PrintSpec() error {
	prettyJSON, err := json.MarshalIndent(oci.spec, "", "  ")
	if err != nil {
		logger.Log.Printf("Failed to marshal JSON, Error: %v", err)
		return err
	}

	fmt.Printf(string(prettyJSON))
	fmt.Printf("\n")

	return nil
}
