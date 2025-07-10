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

package main

import (
	"fmt"
	"os"

	"github.com/ROCm/container-toolkit/cmd/amd-ctk/cdi"
	"github.com/ROCm/container-toolkit/cmd/amd-ctk/runtime"
	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/urfave/cli/v2"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "none"
)

func showVersion() *cli.Command {
	showVersionCmd := cli.Command{
		Name:  "version",
		Usage: "Show the version",
		Action: func(c *cli.Context) error {
			fmt.Printf("Version: %s\nBuild Date: %s\nGitCommit: %s\n", Version, BuildDate, GitCommit)
			return nil
		},
	}

	return &showVersionCmd
}

func main() {
	logger.Init(false)

	// Create the top-level CLI tree
	amdCtkCli := &cli.App{
		Name:                 "AMD Container Toolkit CLI",
		EnableBashCompletion: true,
		Usage:                "Tool to configure AMD Container Toolkit",
	}

	// Add subcommands
	amdCtkCli.Commands = []*cli.Command{
		showVersion(),
		runtime.AddNewCommand(),
		cdi.AddNewCommand(),
	}

	err := amdCtkCli.Run(os.Args)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
