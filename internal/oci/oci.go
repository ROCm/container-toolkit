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
	"strings"

	"github.com/ROCm/container-runtime/internal/logger"
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
	// isCreate specifies if the container is getting created now
	isCreate bool
	// origSpecPath is where the input OCI spec is on the disk
	origSpecPath string
	// updatedSpecPath is where the updated OCI spec is put on the disk
	updatedSpecPath string
	// spec is the structure into which the input spec file is read into
	spec *specs.Spec
	// hookPath is the where the OCI hook executable is on the disk
	hookPath string
}

// SpecUpdateOp specifies type of update operation on the OCI spec
type SpecUpdateOp int

// Enumeration of the update operations on the OCI spec
const (
	AddHook SpecUpdateOp = iota
	AddLinuxDevice
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
				i = i + 1
			}
		} else if args[i] == "create" {
			oci.isCreate = true
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
		logger.Log.Printf("Error opening file:%s", err)
		return err
	}

	defer file.Close()

	var spec specs.Spec
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&spec); err != nil {
		logger.Log.Printf("Failed to decode JSON: %s\n", err)
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

// addLinuxDevice adds Linux device to the OCI spec.
func (oci *oci_t) addLinuxDevice() error {
	fm := 444
	ofm := os.FileMode(fm)

	var gid uint32 = 0

	dev := specs.LinuxDevice{
		Path:     "/dev/dri/renderD128",
		Major:    0,
		Minor:    0,
		FileMode: &ofm,
		GID:      &gid,
	}

	if oci.spec.Linux == nil {
		oci.spec.Linux = &specs.Linux{}
	}

	oci.spec.Linux.Devices = append(oci.spec.Linux.Devices, dev)

	var major, minor int64 = 0, 0
	rdev := specs.LinuxDeviceCgroup{
		Allow:  true,
		Type:   "c",
		Major:  &major,
		Minor:  &minor,
		Access: "rwm",
	}

	if oci.spec.Linux.Resources == nil {
		oci.spec.Linux.Resources = &specs.LinuxResources{}
	}

	oci.spec.Linux.Resources.Devices = append(oci.spec.Linux.Resources.Devices, rdev)

	logger.Log.Printf("Added Linux device %v to OCI spec", "render")

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

	return oci, nil
}

// IsCreate returns true if the container is getting created now
func (oci *oci_t) IsCreate() bool {
	return oci.isCreate
}

// WriteSpec writes the updated spec back to disk
func (oci *oci_t) WriteSpec() error {
	f := oci.updatedSpecPath

	file, err := os.Create(f)
	if err != nil {
		logger.Log.Printf("Error creating file:%s\n", err)
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(oci.spec); err != nil {
		fmt.Printf("Error encoding JSON: %s\n", err)
		return err
	}

	logger.Log.Printf("Wrote the spec to %v", f)
	return nil
}

// UpdateSpec updates the input OCI spec as per the request op
func (oci *oci_t) UpdateSpec(op SpecUpdateOp) error {
	switch op {
	case AddHook:
		return oci.addHook()
	case AddLinuxDevice:
		return oci.addLinuxDevice()
	}

	return nil
}

// PrintSpec prints the current spec on the console
func (oci *oci_t) PrintSpec() error {
	prettyJSON, err := json.MarshalIndent(oci.spec, "", "  ")
	if err != nil {
		logger.Log.Printf("Failed to marshal JSON: %s", err)
		return err
	}

	fmt.Printf(string(prettyJSON))
	fmt.Printf("\n")

	return nil
}
