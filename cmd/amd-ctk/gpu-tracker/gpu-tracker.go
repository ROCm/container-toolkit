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

package gpuTracker

import (
	"fmt"
	"os/user"

	"github.com/ROCm/container-toolkit/cmd/amd-ctk/gpu-tracker/disable"
	"github.com/ROCm/container-toolkit/cmd/amd-ctk/gpu-tracker/enable"
	"github.com/ROCm/container-toolkit/cmd/amd-ctk/gpu-tracker/initialize"
	"github.com/ROCm/container-toolkit/cmd/amd-ctk/gpu-tracker/release"
	"github.com/ROCm/container-toolkit/cmd/amd-ctk/gpu-tracker/reset"
	"github.com/ROCm/container-toolkit/cmd/amd-ctk/gpu-tracker/status"
	"github.com/ROCm/container-toolkit/internal/gpu-tracker"
	"github.com/urfave/cli/v2"
)

func AddNewCommand() *cli.Command {
	gpuTrackerCmd := cli.Command{
		Name:  "gpu-tracker",
		Usage: "GPU Tracker related commands",
		UsageText: `amd-ctk gpu-tracker [gpu-ids] [accessibility]

	Arguments:
		gpu-ids        Comma-separated list of GPU IDs (comma separated list, range operator, all)
		accessibility  Must be either 'exclusive' or 'shared'

	Examples:
		amd-ctk gpu-tracker 0,1,2 exclusive
		amd-ctk gpu-tracker 0,1-2 shared
		amd-ctk gpu-tracker all shared

OR

amd-ctk gpu-tracker [command] [options]`,
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c)
		},
	}

	gpuTrackerCmd.Subcommands = []*cli.Command{
		disable.AddNewCommand(),
		enable.AddNewCommand(),
		initialize.AddNewCommand(),
		reset.AddNewCommand(),
		release.AddNewCommand(),
		status.AddNewCommand(),
	}

	return &gpuTrackerCmd
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

	if c.Args().Len() < 2 {
		return cli.Exit("Error: Missing arguments. Usage: gpu-tracker <gpu_ids> <operation>", 1)
	}

	gpuIDs := c.Args().Get(0)
	operation := c.Args().Get(1)

	gpuTracker, err := gpuTracker.New()
	if err != nil {
		return fmt.Errorf("Failed to create GPU tracker, Error: %v", err)
	}

	switch operation {
	case "exclusive":
		gpuTracker.MakeGPUsExclusive(gpuIDs)
	case "shared":
		gpuTracker.MakeGPUsShared(gpuIDs)
	default:
		return cli.Exit("Error: Invalid operation. Must be 'exclusive' or 'shared'", 1)
	}

	return nil
}
