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

package release

import (
	"fmt"
	"os/user"

	"github.com/ROCm/container-toolkit/internal/gpu-tracker"
	"github.com/urfave/cli/v2"
)

func AddNewCommand() *cli.Command {
	// Add the gpu-tracker release command
	gpuTrackerReleaseCmd := cli.Command{
		Name:   "release",
		Hidden: true,
		Usage:  "Release GPUs used by a container",
		UsageText: `amd-ctk gpu-tracker release [container_id]

	Arguments:
		container_id  container ID of the container

	Examples:
		amd-ctk gpu-tracker release a4e19862b4e2a1b04a1f793f346d0411f4a0a3857578c526a25ac6c858168fd8`,
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c)
		},
	}

	return &gpuTrackerReleaseCmd
}

func validateGenOptions(c *cli.Context) error {
	curUser, err := user.Current()
	if err != nil || curUser.Uid != "0" {
		return fmt.Errorf("Permission denied: Not running as root")
	}

	return nil
}

func performAction(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return cli.ShowAppHelp(c)
	}

	gpuTracker, err := gpuTracker.New()
	if err != nil {
		return fmt.Errorf("Failed to create GPU tracker, Error: %v", err)
	}

	err = gpuTracker.ReleaseGPUs(c.Args().Get(0))
	if err != nil {
		return fmt.Errorf("Failed to release GPUs, Error: %v", err)
	}

	return nil
}
