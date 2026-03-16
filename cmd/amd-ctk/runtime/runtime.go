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

package runtime

import (
	"fmt"
	"os"

	"github.com/ROCm/container-toolkit/cmd/amd-ctk/runtime/configure"
	"github.com/urfave/cli/v2"
)

func AddNewCommand() *cli.Command {
	// Add the runtime command
	runtimeCmd := cli.Command{
		Name:      "runtime",
		Usage:     "runtime related commands for AMD Container Toolkit",
		UsageText: "amd-ctk runtime [command] [options]",
	}

	runtimeCmd.Subcommands = []*cli.Command{
		configure.AddNewCommand(),
		addConfigureHookCommand(),
	}

	return &runtimeCmd
}

func addConfigureHookCommand() *cli.Command {
	return &cli.Command{
		Name:  "configure-hook",
		Usage: "Install amd-container-runtime-hook as OCI hook",
		Action: func(c *cli.Context) error {
			hookPath := "/usr/bin/amd-container-runtime-hook"
			if _, err := os.Stat(hookPath); os.IsNotExist(err) {
				return fmt.Errorf("hook binary not found at %s", hookPath)
			}
			fmt.Printf("AMD Container Runtime Hook is available at: %s\n", hookPath)
			fmt.Println("Add this hook to your runtime configuration to enable --gpus flag support")
			return nil
		},
	}
}
