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

package list

import (
	"fmt"
	"strings"

	"github.com/ROCm/container-toolkit/internal/amdgpu"
	"github.com/urfave/cli/v2"
)

func AddNewCommand() *cli.Command {
	// Add the cdi list command
	cdiListCmd := cli.Command{
		Name:  "list",
		Usage: "List the available AMD GPU devices",
		Action: func(c *cli.Context) error {
			return performAction(c)
		},
	}

	return &cdiListCmd
}

func performAction(c *cli.Context) error {
	devs, err := amdgpu.GetAMDGPUs()
	if err != nil {
		return fmt.Errorf("failed to list AMD devices: %v", err)
	}

	suffix := "devices"
	if len(devs) == 1 {
		suffix = "device"
	}
	fmt.Printf("Found %v AMD GPU %s\n", len(devs), suffix)
	for cnt, dev := range devs {
		if cnt == 0 {
			fmt.Printf("amd.com/gpu=all\n")
		}
		fmt.Printf("amd.com/gpu=%v\n", cnt)
		for _, dd := range dev {
			if !strings.HasPrefix(dd, "/dev/dri/card") {
				fmt.Printf("  %s\n", dd)
			}
		}
	}
	return nil
}
