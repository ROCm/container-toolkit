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
	"os"
	"path/filepath"
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDeviceList(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedDevices []int
		expectError     bool
	}{
		{
			name:            "empty string",
			input:           "",
			expectedDevices: []int{},
			expectError:     false,
		},
		{
			name:            "void",
			input:           "void",
			expectedDevices: []int{},
			expectError:     false,
		},
		{
			name:            "single device index",
			input:           "0",
			expectedDevices: []int{0},
			expectError:     false,
		},
		{
			name:            "multiple device indices",
			input:           "0,1,2",
			expectedDevices: []int{0, 1, 2},
			expectError:     false,
		},
		{
			name:            "invalid device",
			input:           "invalid",
			expectedDevices: nil,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices, err := parseDeviceList(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDevices, devices)
			}
		})
	}
}

func TestExtractGPUDevices(t *testing.T) {
	tests := []struct {
		name            string
		env             []string
		expectedDevices []int
		expectError     bool
	}{
		{
			name:            "no GPU env",
			env:             []string{"PATH=/bin"},
			expectedDevices: []int{},
			expectError:     false,
		},
		{
			name:            "AMD_VISIBLE_DEVICES with indices",
			env:             []string{"AMD_VISIBLE_DEVICES=0,1"},
			expectedDevices: []int{0, 1},
			expectError:     false,
		},
		{
			name:            "DOCKER_RESOURCE_GPU with index",
			env:             []string{"DOCKER_RESOURCE_GPU=0"},
			expectedDevices: []int{0},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &specs.Spec{
				Process: &specs.Process{
					Env: tt.env,
				},
			}

			devices, err := extractGPUDevices(spec)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDevices, devices)
			}
		})
	}
}

func TestLoadSpec(t *testing.T) {
	// Create temporary test bundle
	tmpDir := t.TempDir()
	
	testSpec := specs.Spec{
		Version: stringPtr("1.0.0"),
		Root:    &specs.Root{Path: "/rootfs"},
		Process: &specs.Process{
			Env: []string{"PATH=/bin"},
		},
	}

	specPath := filepath.Join(tmpDir, "config.json")
	data, err := json.Marshal(testSpec)
	require.NoError(t, err)
	
	err = os.WriteFile(specPath, data, 0644)
	require.NoError(t, err)

	// Test loading
	spec, err := loadSpec(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, spec)
	assert.Equal(t, "1.0.0", *spec.Version)
	assert.Equal(t, "/rootfs", spec.Root.Path)
}

func TestValidateGPUDevices(t *testing.T) {
	tests := []struct {
		name           string
		devices        []int
		allowedDevices []int
		expectError    bool
	}{
		{
			name:           "no restrictions",
			devices:        []int{0, 1, 2},
			allowedDevices: []int{},
			expectError:    false,
		},
		{
			name:           "all devices allowed",
			devices:        []int{0, 1},
			allowedDevices: []int{0, 1, 2, 3},
			expectError:    false,
		},
		{
			name:           "device not allowed",
			devices:        []int{2},
			allowedDevices: []int{0, 1},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGPUDevices(tt.devices, tt.allowedDevices)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
