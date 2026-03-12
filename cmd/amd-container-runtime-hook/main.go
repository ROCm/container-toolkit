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
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/ROCm/container-toolkit/internal/gpu-tracker"
	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/opencontainers/runtime-spec/specs-go"
)

var (
	debugflag   = flag.Bool("debug", false, "enable debug output")
	versionflag = flag.Bool("version", false, "enable version output")
	configflag  = flag.String("config", "", "configuration file")
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "none"
)

// HookState holds state information about the hook
type HookState struct {
	Pid int `json:"pid,omitempty"`
	// OCI runtime spec >= 1.0.0
	Bundle string `json:"bundle"`
	// OCI runtime spec < 1.0.0 (runc legacy)
	BundlePath string `json:"bundlePath"`
}

// ContainerConfig represents the configuration extracted from OCI spec
type ContainerConfig struct {
	Pid         int
	Rootfs      string
	GPUDevices  []int
	ContainerID string
}

func exit() {
	if err := recover(); err != nil {
		if _, ok := err.(runtime.Error); ok {
			log.Println(err)
		}
		if *debugflag {
			log.Printf("%s", debug.Stack())
		}
		os.Exit(1)
	}
	os.Exit(0)
}

// loadSpec loads the OCI spec from the bundle directory
func loadSpec(path string) (*specs.Spec, error) {
	f, err := os.Open(filepath.Join(path, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("could not open OCI spec: %v", err)
	}
	defer f.Close()

	var spec specs.Spec
	if err = json.NewDecoder(f).Decode(&spec); err != nil {
		return nil, fmt.Errorf("could not decode OCI spec: %v", err)
	}

	if spec.Version == nil {
		return nil, fmt.Errorf("Version is empty in OCI spec")
	}
	if spec.Process == nil {
		return nil, fmt.Errorf("Process is empty in OCI spec")
	}
	if spec.Root == nil {
		return nil, fmt.Errorf("Root is empty in OCI spec")
	}

	return &spec, nil
}

// getContainerConfig reads container state from stdin and extracts GPU configuration
func getContainerConfig() (*ContainerConfig, error) {
	var state HookState
	if err := json.NewDecoder(os.Stdin).Decode(&state); err != nil {
		return nil, fmt.Errorf("could not decode container state: %v", err)
	}

	bundle := state.Bundle
	if len(bundle) == 0 {
		bundle = state.BundlePath
	}

	spec, err := loadSpec(bundle)
	if err != nil {
		return nil, err
	}

	// Extract GPU devices from environment variables
	gpuDevices, err := extractGPUDevices(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to extract GPU devices: %v", err)
	}

	// Extract container ID from annotations if available
	containerID := ""
	if spec.Annotations != nil {
		if id, ok := spec.Annotations["io.kubernetes.cri.container-id"]; ok {
			containerID = id
		}
	}

	config := &ContainerConfig{
		Pid:         state.Pid,
		Rootfs:      spec.Root.Path,
		GPUDevices:  gpuDevices,
		ContainerID: containerID,
	}

	return config, nil
}

// injectGPUDevices modifies the OCI spec to add requested GPU devices
func injectGPUDevices(bundlePath string, config *ContainerConfig) error {
	spec, err := loadSpec(bundlePath)
	if err != nil {
		return err
	}

	if len(config.GPUDevices) == 0 {
		logger.Log.Printf("No GPU devices to inject")
		return nil
	}

	// Get all AMD GPUs
	devs, err := amdgpu.GetAMDGPUs()
	if err != nil {
		return fmt.Errorf("failed to enumerate AMD GPUs: %v", err)
	}

	// Add requested GPU devices to spec
	if spec.Linux == nil {
		spec.Linux = &specs.Linux{}
	}
	if spec.Linux.Resources == nil {
		spec.Linux.Resources = &specs.LinuxResources{}
	}

	for _, idx := range config.GPUDevices {
		if idx >= len(devs) {
			return fmt.Errorf("invalid GPU index %d (only %d GPUs available)", idx, len(devs))
		}

		for _, drmDev := range devs[idx].DrmDevices {
			gpu, err := amdgpu.GetAMDGPU(drmDev)
			if err != nil {
				logger.Log.Printf("Warning: failed to get device info for %s: %v", drmDev, err)
				continue
			}

			// Add device to spec
			dev := specs.LinuxDevice{
				Path:     gpu.Path,
				Type:     gpu.DevType,
				Major:    gpu.Major,
				Minor:    gpu.Minor,
				FileMode: &gpu.FileMode,
				GID:      &gpu.Gid,
				UID:      &gpu.Uid,
			}
			spec.Linux.Devices = append(spec.Linux.Devices, dev)

			// Add device cgroup rule
			rdev := specs.LinuxDeviceCgroup{
				Allow:  gpu.Allow,
				Type:   gpu.DevType,
				Major:  &gpu.Major,
				Minor:  &gpu.Minor,
				Access: gpu.Access,
			}
			spec.Linux.Resources.Devices = append(spec.Linux.Resources.Devices, rdev)

			logger.Log.Printf("Injected GPU device: %s", gpu.Path)
		}
	}

	// Add /dev/kfd
	kfd, err := amdgpu.GetAMDGPU("/dev/kfd")
	if err != nil {
		return fmt.Errorf("failed to get /dev/kfd info: %v", err)
	}

	dev := specs.LinuxDevice{
		Path:     kfd.Path,
		Type:     kfd.DevType,
		Major:    kfd.Major,
		Minor:    kfd.Minor,
		FileMode: &kfd.FileMode,
		GID:      &kfd.Gid,
		UID:      &kfd.Uid,
	}
	spec.Linux.Devices = append(spec.Linux.Devices, dev)

	rdev := specs.LinuxDeviceCgroup{
		Allow:  kfd.Allow,
		Type:   kfd.DevType,
		Major:  &kfd.Major,
		Minor:  &kfd.Minor,
		Access: kfd.Access,
	}
	spec.Linux.Resources.Devices = append(spec.Linux.Resources.Devices, rdev)

	logger.Log.Printf("Injected /dev/kfd device")

	// Add poststop hook for GPU release
	if spec.Hooks == nil {
		spec.Hooks = &specs.Hooks{}
	}
	releaseHook := specs.Hook{
		Path: "/usr/local/bin/amd-ctk",
		Args: []string{
			"amd-ctk",
			"gpu-tracker",
			"release",
			config.ContainerID,
		},
	}
	spec.Hooks.Poststop = append(spec.Hooks.Poststop, releaseHook)

	// Write modified spec back
	f, err := os.Create(filepath.Join(bundlePath, "config.json"))
	if err != nil {
		return fmt.Errorf("failed to create spec file: %v", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(spec); err != nil {
		return fmt.Errorf("failed to encode spec: %v", err)
	}

	return nil
}

func doPrestart() {
	defer exit()
	log.SetFlags(0)

	// Decode state from stdin first to get bundle path
	var state HookState
	stdinData, err := os.ReadFile("/dev/stdin")
	if err != nil {
		log.Panicf("Failed to read stdin: %v", err)
	}

	if err := json.Unmarshal(stdinData, &state); err != nil {
		log.Panicf("Failed to decode container state: %v", err)
	}

	bundle := state.Bundle
	if len(bundle) == 0 {
		bundle = state.BundlePath
	}

	// Now load and parse the full config
	spec, err := loadSpec(bundle)
	if err != nil {
		log.Panicf("Failed to load OCI spec: %v", err)
	}

	// Extract GPU devices
	gpuDevices, err := extractGPUDevices(spec)
	if err != nil {
		log.Panicf("Failed to extract GPU devices: %v", err)
	}

	if len(gpuDevices) == 0 {
		logger.Log.Printf("No GPUs requested, skipping device injection")
		return
	}

	// Extract container ID
	containerID := ""
	if spec.Annotations != nil {
		if id, ok := spec.Annotations["io.kubernetes.cri.container-id"]; ok {
			containerID = id
		}
	}

	config := &ContainerConfig{
		Pid:         state.Pid,
		Rootfs:      spec.Root.Path,
		GPUDevices:  gpuDevices,
		ContainerID: containerID,
	}

	if err := injectGPUDevices(bundle, config); err != nil {
		log.Panicf("Failed to inject GPU devices: %v", err)
	}

	logger.Log.Printf("Successfully injected %d GPU device(s)", len(config.GPUDevices))
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of amd-container-runtime-hook:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  prestart\n        run the prestart hook\n")
	fmt.Fprintf(os.Stderr, "  poststart\n        no-op\n")
	fmt.Fprintf(os.Stderr, "  poststop\n        no-op\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	logger.Init(*debugflag)

	if *versionflag {
		fmt.Printf("amd-container-runtime-hook version %s\n", Version)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	switch args[0] {
	case "prestart":
		doPrestart()
		os.Exit(0)
	case "poststart":
		fallthrough
	case "poststop":
		os.Exit(0)
	default:
		flag.Usage()
		os.Exit(2)
	}
}
