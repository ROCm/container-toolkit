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
	"fmt"
	"strconv"
	"strings"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// parseDeviceList parses AMD_VISIBLE_DEVICES or DOCKER_RESOURCE_* values
// Returns list of device indices
func parseDeviceList(deviceStr string) ([]int, error) {
	if deviceStr == "" || deviceStr == "void" {
		return []int{}, nil
	}

	if deviceStr == "all" || deviceStr == "All" || deviceStr == "ALL" {
		devs, err := amdgpu.GetAMDGPUs()
		if err != nil {
			return nil, fmt.Errorf("failed to enumerate GPUs: %v", err)
		}
		devices := make([]int, len(devs))
		for i := range devs {
			devices[i] = i
		}
		return devices, nil
	}

	// Parse comma-separated list of indices or UUIDs
	devices := []int{}
	uuidMap, err := amdgpu.GetUniqueIdToDeviceIndexMap()
	if err != nil {
		logger.Log.Printf("Warning: failed to get UUID mappings: %v", err)
		uuidMap = make(map[string][]int)
	}

	for _, part := range strings.Split(deviceStr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if it's a UUID (hex string)
		if strings.HasPrefix(part, "0x") || strings.HasPrefix(part, "0X") {
			uuid := strings.ToLower(part)
			if indices, found := uuidMap[uuid]; found {
				devices = append(devices, indices...)
			} else {
				// Try without 0x prefix
				uuid = strings.TrimPrefix(uuid, "0x")
				if indices, found := uuidMap[uuid]; found {
					devices = append(devices, indices...)
				} else {
					return nil, fmt.Errorf("unknown GPU UUID: %s", part)
				}
			}
		} else if idx, err := strconv.Atoi(part); err == nil {
			// It's a device index
			devices = append(devices, idx)
		} else {
			// Maybe it's a UUID without 0x prefix
			uuid := strings.ToLower(part)
			if indices, found := uuidMap[uuid]; found {
				devices = append(devices, indices...)
			} else {
				return nil, fmt.Errorf("invalid device specifier: %s", part)
			}
		}
	}

	return devices, nil
}

// extractGPUDevices extracts requested GPU devices from OCI spec environment
func extractGPUDevices(spec *specs.Spec) ([]int, error) {
	if spec == nil || spec.Process == nil {
		return []int{}, nil
	}

	for _, env := range spec.Process.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Check for AMD_VISIBLE_DEVICES or DOCKER_RESOURCE_*
		if key == "AMD_VISIBLE_DEVICES" || strings.HasPrefix(key, "DOCKER_RESOURCE_") {
			return parseDeviceList(value)
		}
	}

	return []int{}, nil
}

// validateGPUDevices checks if requested devices are within allowed range
func validateGPUDevices(devices []int, allowedDevices []int) error {
	if len(allowedDevices) == 0 {
		// No restrictions
		return nil
	}

	allowed := make(map[int]bool)
	for _, dev := range allowedDevices {
		allowed[dev] = true
	}

	for _, dev := range devices {
		if !allowed[dev] {
			return fmt.Errorf("GPU device %d is not in allowed list: %v", dev, allowedDevices)
		}
	}

	return nil
}
