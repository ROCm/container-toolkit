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
	"os"

	"github.com/pelletier/go-toml"
)

const (
	defaultConfigPath = "/etc/amd-container-runtime/config.toml"
)

// HookConfig represents the configuration for the AMD runtime hook
type HookConfig struct {
	// DebugFilePath specifies where to write debug logs
	DebugFilePath string `toml:"debug-file-path"`
	
	// DisableGPUEnumeration disables automatic GPU discovery
	DisableGPUEnumeration bool `toml:"disable-gpu-enumeration"`
	
	// AllowedDevices restricts which GPU devices can be exposed (empty = all)
	AllowedDevices []int `toml:"allowed-devices"`
}

// loadConfig loads hook configuration from the specified path
func loadConfig(path string) (*HookConfig, error) {
	if path == "" {
		path = defaultConfigPath
	}

	// Default configuration
	config := &HookConfig{
		DebugFilePath:         "",
		DisableGPUEnumeration: false,
		AllowedDevices:        []int{},
	}

	// If config file doesn't exist, return defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, nil
	}

	// Load from file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := toml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

// getConfigFilePath returns the config file path from flag or environment
func getConfigFilePath() string {
	if configflag != nil && *configflag != "" {
		return *configflag
	}
	if env := os.Getenv("AMD_RUNTIME_HOOK_CONFIG"); env != "" {
		return env
	}
	return defaultConfigPath
}
