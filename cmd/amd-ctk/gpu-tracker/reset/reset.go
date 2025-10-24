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

package reset

import (
	"fmt"
	"os/user"

	"github.com/ROCm/container-toolkit/internal/gpu-tracker"
	"github.com/urfave/cli/v2"
)

func AddNewCommand() *cli.Command {
	// Add the gpu-tracker reset command
	gpuTrackerResetCmd := cli.Command{
		Name:      "reset",
		Usage:     "Reset the GPU Tracker",
		UsageText: "amd-ctk gpu-tracker reset [options]",
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c)
		},
	}

	return &gpuTrackerResetCmd
}

func validateGenOptions(c *cli.Context) error {
	curUser, err := user.Current()
	if err != nil || curUser.Uid != "0" {
		return fmt.Errorf("Permission denied: Not running as root")
	}

	return nil
}

func performAction(c *cli.Context) error {
	gpuTracker, err := gpuTracker.New()
	if err != nil {
		return fmt.Errorf("Failed to create GPU tracker, Error: %v", err)
	}

	err = gpuTracker.Reset()
	if err != nil {
		return fmt.Errorf("Failed to Reset GPU Tracker, Error: %v", err)
	}

	return nil
}
