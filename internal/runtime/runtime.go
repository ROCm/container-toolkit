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

package runtime

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	"github.com/ROCm/container-toolkit/internal/oci"
)

// Constants
const (
	// runc executable
	RUNC = "runc"
)

// Interface for runtime package
type Interface interface {
	// Run starts the runtime
	Run() error
}

// runtm implements the runtime interface
type runtm struct {
	// args is the argument list passed to runtime
	args []string
	// oci is the handle for oci operations
	oci oci.Interface
}

// New creates a runtime instance
func New(args []string) (Interface, error) {
	var err error

	rt := &runtm{
		args: args,
	}

	rt.oci, err = oci.New(rt.args[1:])
	if err != nil {
		return nil, fmt.Errorf("creating OCI handler: %w", err)
	}

	return rt, nil
}

// Run starts the runtime
func (rt *runtm) Run() error {
	var err error

	if rt.oci.HasHelpOption() {
		fmt.Printf("\nAMD Container Runtime is a wrapper over runc. Below is the help for runc.\n\n")
	}

	if rt.oci.IsCreate() {
		// Add GPUs
		err = rt.oci.UpdateSpec(oci.AddGPUDevices)
		if err != nil {
			return fmt.Errorf("update OCI spec (add GPU devices): %w", err)
		}

		/*
			// Print updated OCI spec
			err = rt.oci.PrintSpec()
			if err != nil {
				return err
			}
		*/

		// Write updated OCI spec
		err = rt.oci.WriteSpec()
		if err != nil {
			return fmt.Errorf("write OCI spec: %w", err)
		}
		slog.Info("Container configured for GPU access")
	}

	// Call runc with updated oci spec
	runc, err := exec.LookPath(RUNC)
	if err != nil {
		return fmt.Errorf("unable to find runc in PATH: %w", err)
	}

	slog.Info("Launching container")
	slog.Debug("Running runc", "args", rt.args, "environ", os.Environ())
	err = syscall.Exec(runc, rt.args, os.Environ())
	if err != nil {
		return fmt.Errorf("calling runc: %w", err)
	}

	return nil
}
