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
	gpuTrackerLib "github.com/ROCm/container-toolkit/internal/gpu-tracker"
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

	gpuTracker, err := gpuTrackerLib.New()
	if err != nil {
		return fmt.Errorf("failed to create GPU Tracker: %w", err)
	}

	enabled, err := gpuTracker.IsEnabled()
	if err != nil {
		return fmt.Errorf("failed to check GPU Tracker status: %w", err)
	}
	if !enabled {
		fmt.Println("GPU Tracker is disabled")
		return nil
	}

	var res *gpuTrackerLib.AccessibilityResult
	switch operation {
	case "exclusive":
		res, err = gpuTracker.MakeGPUsExclusive(gpuIDs)
		if err != nil {
			return fmt.Errorf("failed to make GPUs %s exclusive: %w", gpuIDs, err)
		}
		if len(res.Changed) > 0 {
			fmt.Printf("GPUs %v have been made exclusive\n", res.Changed)
		}
		if len(res.NotChanged) > 0 {
			fmt.Printf("GPUs %v have not been made exclusive because more than one container is currently using it\n", res.NotChanged)
		}
	case "shared":
		res, err = gpuTracker.MakeGPUsShared(gpuIDs)
		if err != nil {
			return fmt.Errorf("failed to make GPUs %s shared: %w", gpuIDs, err)
		}
		if len(res.Changed) > 0 {
			fmt.Printf("GPUs %v have been made shared\n", res.Changed)
		}
	default:
		return cli.Exit("Error: Invalid operation. Must be 'exclusive' or 'shared'", 1)
	}

	if len(res.InvalidRanges) > 0 {
		fmt.Printf("Ignoring %v GPUs Ranges as they are invalid\n", res.InvalidRanges)
	}
	if len(res.InvalidGPUs) > 0 {
		fmt.Printf("Ignoring %v GPUs as they are invalid\n", res.InvalidGPUs)
	}

	return nil
}
