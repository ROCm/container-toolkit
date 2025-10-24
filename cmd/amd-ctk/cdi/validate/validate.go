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

package validate

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/ROCm/container-toolkit/internal/cdi"
	"github.com/urfave/cli/v2"
)

const (
	defaultCDISpecPath = "/etc/cdi"
	defaultCDISpecFile = "amd.json"
)

type validateOptions struct {
	cdiSpecPath string
}

func AddNewCommand() *cli.Command {
	// Add the cdi validate command
	valOptions := validateOptions{}
	cdiValidateCmd := cli.Command{
		Name:      "validate",
		Usage:     "Validate the CDI spec for GPUs",
		UsageText: "amd-ctk cdi validate [options]",
		Before: func(c *cli.Context) error {
			return validateValOptions(c, &valOptions)
		},
		Action: func(c *cli.Context) error {
			return performAction(c, &valOptions)
		},
	}

	cdiValidateCmd.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "path",
			Usage:       "full path of CDI spec file",
			Value:       defaultCDISpecPath + "/" + defaultCDISpecFile,
			Destination: &valOptions.cdiSpecPath,
		},
	}

	return &cdiValidateCmd
}

func validateValOptions(c *cli.Context, valOptions *validateOptions) error {
	curUser, err := user.Current()
	if err != nil || curUser.Uid != "0" {
		return fmt.Errorf("Permission denied: Not running as root")
	}

	out, err := filepath.Abs(valOptions.cdiSpecPath)
	if err != nil {
		return fmt.Errorf("Incorrect CDI spec file, Error: %v", err)
	}

	f := filepath.Base(out)
	if f != defaultCDISpecFile {
		return fmt.Errorf("CDI spec file name must be amd.json")
	}

	valOptions.cdiSpecPath = strings.TrimSuffix(out, f)
	return nil
}

func performAction(c *cli.Context, valOptions *validateOptions) error {
	cdi, err := cdi.New(valOptions.cdiSpecPath)
	if err != nil {
		return fmt.Errorf("Failed to create CDI handler, Error: %v", err)
	}

	// Validate CDI spec
	result, err := cdi.ValidateSpec()
	if err != nil {
		return fmt.Errorf("Failed to validate CDI spec")
	}
	if result == true {
		fmt.Printf("CDI spec is valid\n")
	}

	return nil
}
