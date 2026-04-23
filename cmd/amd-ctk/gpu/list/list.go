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

package list

import (
	"fmt"
	"strings"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/urfave/cli/v2"
)

func AddNewCommand() *cli.Command {
	gpuListCmd := cli.Command{
		Name:      "list",
		Usage:     "List AMD GPUs with their UUIDs",
		UsageText: "amd-ctk gpu list",
		Action: func(c *cli.Context) error {
			return performAction(c)
		},
	}

	return &gpuListCmd
}

func performAction(c *cli.Context) error {
	devs, err := amdgpu.GetAMDGPUs()
	if err != nil {
		return fmt.Errorf("failed to list AMD devices: %v", err)
	}

	uuidToGPUIdMap, err := amdgpu.GetUniqueIdToDeviceIndexMap()
	if err != nil {
		uuidToGPUIdMap = make(map[string][]int)
	}

	gpuIdToUUIDMap := make(map[int]string)
	for uuid, gpuIds := range uuidToGPUIdMap {
		if strings.HasPrefix(uuid, "0x") || strings.HasPrefix(uuid, "0X") {
			uuid = uuid[2:]
		}
		uuid = "0x" + strings.ToUpper(uuid)
		for _, gpuId := range gpuIds {
			gpuIdToUUIDMap[gpuId] = uuid
		}
	}

	suffix := "devices"
	if len(devs) == 1 {
		suffix = "device"
	}
	fmt.Printf("Found %v AMD GPU %s\n", len(devs), suffix)

	fmt.Println(strings.Repeat("-", 75))
	fmt.Printf("%-10s%-25s%-40s\n", "GPU Id", "UUID", "DRM Devices")
	fmt.Println(strings.Repeat("-", 75))
	for idx, dev := range devs {
		uuid := gpuIdToUUIDMap[idx]
		if uuid == "" {
			uuid = "N/A"
		}

		var renderDevs []string
		for _, dd := range dev.DrmDevices {
			if !strings.HasPrefix(dd, "/dev/dri/card") {
				renderDevs = append(renderDevs, dd)
			}
		}

		drmStr := strings.Join(renderDevs, ", ")
		fmt.Printf("%-10v%-25s%-40s\n", idx, uuid, drmStr)
	}

	return nil
}
