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

package generate

import (
	"fmt"
	"path/filepath"

	"github.com/ROCm/container-toolkit/internal/cdi"
	"github.com/urfave/cli/v2"
)

const (
	defaultOutputFile = "/etc/cdi/amd.json"
)

type generateOptions struct {
	output string
}

func AddNewCommand() *cli.Command {
	// Add the cdi generate command
	genOptions := generateOptions{}
	cdiGenerateCmd := cli.Command{
		Name:  "generate",
		Usage: "Generate the CDI spec for GPUs",
		Before: func(c *cli.Context) error {
			return validateGenOptions(c, &genOptions)
		},
		Action: func(c *cli.Context) error {
			return performAction(c, &genOptions)
		},
	}

	cdiGenerateCmd.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "output",
			Usage:       "full path of output file",
			Value:       defaultOutputFile,
			Destination: &genOptions.output,
		},
	}

	return &cdiGenerateCmd
}

func validateGenOptions(c *cli.Context, genOptions *generateOptions) error {
	out, err := filepath.Abs(genOptions.output)
	if err != nil {
		return fmt.Errorf("incorrect output file, Err: %v", err)

	}

	genOptions.output = out
	return nil
}

func performAction(c *cli.Context, genOptions *generateOptions) error {
	cdi, err := cdi.New(genOptions.output)
	if err != nil {
		return fmt.Errorf("Failed to create CDI handler, Error: %v", err)
	}

	// Generate CDI spec
	err = cdi.GenerateSpec()
	if err != nil {
		return fmt.Errorf("Failed to generate CDI spec, Error: %v", err)
	}

	// Write updated CDI spec
	err = cdi.WriteSpec()
	if err != nil {
		return fmt.Errorf("Failed to write generated runtime CDI spec, Error: %v", err)
	}

	return nil
}
