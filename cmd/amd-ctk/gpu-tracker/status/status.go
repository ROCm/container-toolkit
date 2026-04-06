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

package status

import (
	"fmt"
	"os/user"
	"strings"

	gpuTracker "github.com/ROCm/container-toolkit/internal/gpu-tracker"
	"github.com/urfave/cli/v2"
)

func AddNewCommand() *cli.Command {
	// Add the gpu-tracker status command
	gpuTrackerStatusCmd := cli.Command{
		Name:      "status",
		Usage:     "Show Status of GPUs",
		UsageText: "amd-ctk gpu-tracker status [options]",
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c)
		},
	}

	return &gpuTrackerStatusCmd
}

func validateGenOptions(c *cli.Context) error {
	curUser, err := user.Current()
	if err != nil || curUser.Uid != "0" {
		return fmt.Errorf("Permission denied: Not running as root")
	}

	return nil
}

func performAction(c *cli.Context) error {
	tracker, err := gpuTracker.New()
	if err != nil {
		return fmt.Errorf("creating GPU tracker: %w", err)
	}

	enabled, err := tracker.IsEnabled()
	if err != nil {
		return fmt.Errorf("checking GPU Tracker status: %w", err)
	}
	if !enabled {
		fmt.Println("GPU Tracker is disabled")
		return nil
	}

	entries, err := tracker.ShowStatus()
	if err != nil {
		return fmt.Errorf("showing GPU status: %w", err)
	}

	fmt.Println(strings.Repeat("-", 120))
	fmt.Printf("%-10s%-25s%-20s%-65s\n", "GPU Id", "UUID", "Accessibility", "Container Ids")
	fmt.Println(strings.Repeat("-", 120))
	for _, entry := range entries {
		if len(entry.ContainerIds) > 0 {
			for idx, id := range entry.ContainerIds {
				if idx == 0 {
					fmt.Printf("%-10v%-25v%-20v%-65v\n", entry.GPUId, entry.UUID, entry.Accessibility, id)
				} else {
					fmt.Printf("%-10v%-25v%-20v%-65v\n", "", "", "", id)
				}
			}
		} else {
			fmt.Printf("%-10v%-25v%-20v%-65v\n", entry.GPUId, entry.UUID, entry.Accessibility, "-")
		}
	}

	return nil
}
