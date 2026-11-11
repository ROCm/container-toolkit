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

package remove

import (
	"errors"
	"fmt"

	"github.com/ROCm/container-toolkit/internal/cdi"
	"github.com/urfave/cli/v2"
)

const (
	defaultCDISpecPath = "/etc/cdi"
	defaultCDISpecFile = "amd.json"
)

type removeOptions struct {
	cdiSpecPath string
}

func AddNewCommand() *cli.Command {
	// Add the cdi remove command
	rmOptions := removeOptions{}
	cdiRemoveCmd := cli.Command{
		Name:      "remove",
		Usage:     "Remove the CDI spec for GPUs",
		UsageText: "amd-ctk cdi remove [options]",
		Action: func(c *cli.Context) error {
			return performAction(c, &rmOptions)
		},
	}

	cdiRemoveCmd.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "path",
			Usage:       "full path of CDI spec file",
			Value:       defaultCDISpecPath + "/" + defaultCDISpecFile,
			Destination: &rmOptions.cdiSpecPath,
		},
	}

	return &cdiRemoveCmd
}

func performAction(c *cli.Context, rmOptions *removeOptions) error {
	removed, err := cdi.RemoveSpec(rmOptions.cdiSpecPath)
	if errors.Is(err, cdi.ErrNotAMDSpec) {
		fmt.Printf("Not removing %v: it is not an AMD CDI spec\n", rmOptions.cdiSpecPath)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to remove CDI spec: %w", err)
	}

	if removed {
		fmt.Printf("Removed CDI spec: %v\n", rmOptions.cdiSpecPath)
	} else {
		fmt.Printf("No CDI spec found at %v\n", rmOptions.cdiSpecPath)
	}
	return nil
}
