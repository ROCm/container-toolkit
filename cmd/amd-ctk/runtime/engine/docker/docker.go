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

package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const (
	runtimesKey       = "runtimes"
	defaultRuntimeKey = "default-runtime"
	featuresKey       = "features"
)

type dockerConfig map[string]interface{}

func New(path string) (*dockerConfig, error) {
	return loadConfigFile(path)
}

func loadConfigFile(path string) (*dockerConfig, error) {
	//check if the file exists
	f, err := os.Stat(path)
	if os.IsExist(err) && f.IsDir() {
		return nil, fmt.Errorf("file path is a directory")
	}

	config := dockerConfig{}

	if os.IsNotExist(err) {
		// return empty config
		return &config, nil
	}

	fmt.Printf("Loading configuration from: %v\n", path)

	readB, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %v | err: %v", path, err)
	}

	reader := bytes.NewReader(readB)
	err = json.NewDecoder(reader).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("error decoding configuration file: %v | err: %v", path, err)
	}

	return &config, nil
}

func (d *dockerConfig) ConfigRuntime(name string, path string, isDefault bool) error {
	if d == nil {
		return fmt.Errorf("configuration is empty")
	}

	currentCfg := *d

	//check any existing "runtimes"
	runtimes := map[string]interface{}{}
	if _, exists := currentCfg[runtimesKey]; exists {
		runtimes = currentCfg[runtimesKey].(map[string]interface{})
	}

	runtimes[name] = map[string]interface{}{
		"path": path,
		"args": []string{},
	}
	currentCfg[runtimesKey] = runtimes

	// Enable CDI by default
	currentCfg[featuresKey] = map[string]interface{}{
		"cdi": true,
	}

	if isDefault {
		currentCfg[defaultRuntimeKey] = name
	}

	*d = currentCfg
	return nil
}

func (d *dockerConfig) UnsetDefaultRuntime() error {
	if d == nil {
		return fmt.Errorf("configuration is empty")
	}

	currentCfg := *d

	delete(currentCfg, defaultRuntimeKey)

	fmt.Println("Removed amd as the default runtime")
	return nil
}

func (d *dockerConfig) RemoveRuntime(name string) error {
	if d == nil {
		return fmt.Errorf("configuration is empty")
	}

	currentCfg := *d

	//check any existing "runtimes"
	if _, exists := currentCfg[runtimesKey]; exists {
		runtimes := currentCfg[runtimesKey].(map[string]interface{})
		delete(runtimes, name)
		delete(currentCfg, featuresKey)
		delete(currentCfg, defaultRuntimeKey)
		if len(runtimes) == 0 {
			delete(currentCfg, runtimesKey)
		}
	}

	*d = currentCfg
	return nil
}

func (d dockerConfig) Update(path string) (int, error) {
	toWrite, err := json.MarshalIndent(d, "", "    ")

	if err != nil {
		return 0, fmt.Errorf("json marshal failed: %v", err)
	}
	if path == "" {
		num, err := os.Stdout.Write(toWrite)
		return num, err
	}

	f, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %v | err: %v", path, err)
	}
	defer f.Close()
	return f.Write(toWrite)
}
