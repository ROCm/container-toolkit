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

package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/ROCm/container-toolkit/internal/cdi"
	"github.com/ROCm/container-toolkit/internal/logger"
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
	// cdi is the handle for cdi operations
	cdi cdi.Interface
}

// New creates a runtime instance
func New(args []string) (Interface, error) {
	var err error

	rt := &runtm{
		args: args,
	}

	rt.oci, err = oci.New(rt.args[1:])
	if err != nil {
		logger.Log.Printf("Failed to create OCI handler, Error: %v", err)
		return nil, err
	}

	rt.cdi, err = cdi.New("")
	if err != nil {
		logger.Log.Printf("Failed to create CDI handler, Error: %v", err)
		return nil, err
	}

	return rt, nil
}

// Run starts the runtime
func (rt *runtm) Run() error {
	var err error

	// Generate CDI spec
	err = rt.cdi.GenerateSpec()
	if err != nil {
		logger.Log.Printf("Failed to generate CDI spec, Error: %v", err)
		return err
	}

	// Write updated CDI spec
	err = rt.cdi.WriteSpec()
	if err != nil {
		logger.Log.Printf("Failed to write generated runtime CDI spec, Error: %v", err)
		return err
	}

	if rt.oci.HasHelpOption() {
		fmt.Printf("\nAMD Container Runtime is a wrapper over runc. Below is the help for runc.\n\n")
	}

	if rt.oci.IsCreate() {
		// Add GPUs
		err = rt.oci.UpdateSpec(oci.AddGPUDevices)
		if err != nil {
			logger.Log.Printf("Failed to add Linux device into OCI spec, Error: %v", err)
			return err
		}

		/*
			// Print updated OCI spec
			err = rt.oci.PrintSpec()
			if err != nil {
				logger.Log.Printf("Failed to print runtime OCI spec, Error: %v", err)
				return err
			}
		*/

		// Write updated OCI spec
		err = rt.oci.WriteSpec()
		if err != nil {
			logger.Log.Printf("Failed to write updated runtime OCI spec, Error: %v", err)
			return err
		}
	}

	// Call runc with updated oci spec
	runc, err := exec.LookPath(RUNC)
	if err != nil {
		logger.Log.Printf("Unable to find runc in PATH, Error: %v", err)
		return err
	}

	logger.Log.Printf("Running runc with args: %v, environ: %v", rt.args, os.Environ())
	err = syscall.Exec(runc, rt.args, os.Environ())
	if err != nil {
		logger.Log.Printf("Failed to call runc, Error: %v", err)
		return err
	}

	return nil
}
