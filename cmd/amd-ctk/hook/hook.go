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

package hook

import (
	"github.com/ROCm/container-toolkit/cmd/amd-ctk/hook/create-symlinks"
	"github.com/urfave/cli/v2"
)

func AddNewCommand() *cli.Command {
	// Add the hook command
	hookCmd := cli.Command{
		Name:      "hook",
		Usage:     "OCI hook related commands",
		UsageText: "amd-ctk hook [command] [options]",
	}

	hookCmd.Subcommands = []*cli.Command{
		symlinks.AddNewCommand(),
	}

	return &hookCmd
}
