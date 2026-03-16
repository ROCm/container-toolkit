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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	gpuTracker "github.com/ROCm/container-toolkit/internal/gpu-tracker"
	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/ROCm/container-toolkit/internal/oci"
)

func doPrestart() error {
	logger.Log.Println("Running prestart hook")

	// Read hook state from stdin (Docker/containerd provides this)
	hookState, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read hook state from stdin: %v", err)
	}

	// Create OCI interface for hook context
	ociInterface, err := oci.NewFromStdin()
	if err != nil {
		return fmt.Errorf("failed to create OCI interface: %v", err)
	}

	// Load spec from bundle path in hook state
	ociImpl, ok := ociInterface.(*oci.oci_t)
	if !ok {
		return fmt.Errorf("failed to cast OCI interface to oci_t")
	}

	if err := ociImpl.LoadSpecFromHookState(hookState); err != nil {
		return fmt.Errorf("failed to load spec from hook state: %v", err)
	}

	// Check if GPU devices are requested
	spec := ociInterface.GetSpec()
	if spec == nil || spec.Process == nil {
		logger.Log.Println("No process spec found, skipping GPU configuration")
		return nil
	}

	hasGPURequest := false
	for _, env := range spec.Process.Env {
		if strings.HasPrefix(env, "AMD_VISIBLE_DEVICES=") ||
			strings.HasPrefix(env, "DOCKER_RESOURCE_") {
			hasGPURequest = true
			break
		}
	}

	if !hasGPURequest {
		logger.Log.Println("No GPU devices requested, skipping configuration")
		return nil
	}

	// Add GPU devices to spec
	if err := ociInterface.UpdateSpec(oci.AddGPUDevices); err != nil {
		return fmt.Errorf("failed to add GPU devices: %v", err)
	}

	// Write updated spec back
	if err := ociInterface.WriteSpec(); err != nil {
		return fmt.Errorf("failed to write updated spec: %v", err)
	}

	logger.Log.Println("Successfully configured GPU devices")
	return nil
}

func doPoststop() error {
	logger.Log.Println("Running poststop hook")

	// Read hook state to get container ID
	hookState, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read hook state from stdin: %v", err)
	}

	var state struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(hookState, &state); err != nil {
		return fmt.Errorf("failed to parse hook state: %v", err)
	}

	// Release GPUs via tracker
	tracker, err := gpuTracker.New()
	if err != nil {
		return fmt.Errorf("failed to create GPU tracker: %v", err)
	}

	tracker.ReleaseGPUs(state.ID)
	logger.Log.Printf("Released GPUs for container %s", state.ID)
	return nil
}
