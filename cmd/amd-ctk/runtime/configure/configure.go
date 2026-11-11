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

package configure

import (
	"fmt"

	"github.com/ROCm/container-toolkit/cmd/amd-ctk/runtime/engine"
	"github.com/ROCm/container-toolkit/cmd/amd-ctk/runtime/engine/docker"
	"github.com/urfave/cli/v2"
)

const (
	defaultRuntime              = "docker"
	defaultAmdRuntimeName       = "amd"
	defaultAmdRuntimeExecutable = "amd-container-runtime"
	defaultDockerConfigFilePath = "/etc/docker/daemon.json"
)

type configOptions struct {
	runtime        string
	configFilepath string
	setAsDefault   bool
	unSetAsDefault bool
	remove         bool
}

func AddNewCommand() *cli.Command {
	cfgOptions := configOptions{}

	// Add the configure subcommand
	configureCmd := cli.Command{
		Name:      "configure",
		Usage:     "Configure a runtime to the container engine",
		UsageText: "amd-ctk runtime configure [options]",
		Before: func(c *cli.Context) error {
			return validateConfigOptions(c, &cfgOptions)
		},
		Action: func(c *cli.Context) error {
			return performAction(c, &cfgOptions)
		},
	}

	configureCmd.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "runtime",
			Usage:       "target runtime engine, [docker for now]",
			Value:       defaultRuntime,
			Destination: &cfgOptions.runtime,
		},
		&cli.BoolFlag{
			Name:        "remove",
			Usage:       "remove from target runtimes",
			Destination: &cfgOptions.remove,
		},
		&cli.StringFlag{
			Name:        "config-path",
			Usage:       "path to the configuration file for the target engine",
			Value:       defaultDockerConfigFilePath,
			Destination: &cfgOptions.configFilepath,
		},
		&cli.BoolFlag{
			Name:        "amd-set-as-default",
			Aliases:     []string{"set-as-default"},
			Usage:       "set AMD runtime as the default",
			Destination: &cfgOptions.setAsDefault,
		},
		&cli.BoolFlag{
			Name:        "unset-amd-as-default",
			Aliases:     []string{"unset-as-default"},
			Usage:       "remove AMD runtime as the default",
			Destination: &cfgOptions.unSetAsDefault,
		},
	}
	return &configureCmd
}

func validateConfigOptions(c *cli.Context, cfgOptions *configOptions) error {
	if cfgOptions.runtime != "docker" {
		return fmt.Errorf("unsupported runtime engine: %v", cfgOptions.runtime)
	}
	if cfgOptions.setAsDefault && cfgOptions.unSetAsDefault {
		return fmt.Errorf("both set and unset as default cannot be used at the same time")
	}
	if cfgOptions.remove {
		if cfgOptions.setAsDefault || cfgOptions.unSetAsDefault {
			return fmt.Errorf("remove flag cannot be used along with set-as-default flag")
		}
	}
	return nil
}

func performAction(c *cli.Context, cfgOptions *configOptions) error {
	var (
		err           error
		runtimeEngine engine.Interface
		doNotUpdate   bool
	)

	switch cfgOptions.runtime {
	case "docker":
		runtimeEngine, err = docker.New(cfgOptions.configFilepath)
	default:
		return fmt.Errorf("unsupported runtime engine: %v", cfgOptions.runtime)
	}

	if err != nil || runtimeEngine == nil {
		return fmt.Errorf("failed to init config for runtime engine: %v | err: %v", cfgOptions.runtime, err)
	}

	if cfgOptions.unSetAsDefault {
		err = runtimeEngine.UnsetDefaultRuntime()

		if err != nil {
			return fmt.Errorf("failed to unset as default runtime %v", err)
		}
	} else {
		if cfgOptions.remove {
			err, doNotUpdate = runtimeEngine.RemoveRuntime(defaultAmdRuntimeName)
		} else {
			err = runtimeEngine.ConfigRuntime(defaultAmdRuntimeName, defaultAmdRuntimeExecutable, cfgOptions.setAsDefault)
		}

		if err != nil {
			return fmt.Errorf("failed to update configuration: %v", err)
		}
	}

	// Save the config
	if !doNotUpdate {
		num, err := runtimeEngine.Update(cfgOptions.configFilepath)
		if err != nil {
			return fmt.Errorf("failed to save the config: %v", err)
		}

		if num != 0 {
			fmt.Printf("Updated the config file: %v\n", cfgOptions.configFilepath)
		}
		fmt.Printf("Please restart %v daemon\n", cfgOptions.runtime)
	}
	return nil
}
