package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const (
	configFile         = "/tmp/testConfig.json"
	amdRuntimePath     = "amd-container-runtime"
	amdRuntimeName     = "amd"
	removeAndDefErrMsg = "remove flag cannot be used along with set-as-default flag"
	setUnsetDefErrMsg  = "both set and unset as default cannot be used at the same time"
)

var cliPath = os.Getenv("AMD_CTK_PATH")

// Helper function to run the CLI command and return the output/error
func runCLI(args ...string) (string, string, error) {
	cmd := exec.Command(cliPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func setup(t *testing.T) {
	if cliPath == "" {
		t.Fatalf("cliPath is not set, usage: export AMD_CTK_PATH=<path to amd-ctk executable>; go test -v")
	}

	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		t.Fatalf("amd-ctk is not built, please run 'make container-toolkit-ctk'")
	}
}

func verifyConfigFile(t *testing.T, isDefault bool, isEmpty bool) {
	type features struct {
		Cdi bool `json:"cdi"`
	}
	type runtimeConfig struct {
		Args []string `json:"args"`
		Path string   `json:"path"`
	}
	type runtimes map[string]runtimeConfig

	type config struct {
		DefaultRuntime string   `json:"default-runtime"`
		Features       features `json:"features"`
		Runtimes       runtimes `json:"runtimes"`
	}

	// read the configFile
	_, err := os.Stat(configFile)

	Assert(t, os.IsNotExist(err) == false, fmt.Sprintf("config file: %v doesn't exist", configFile))
	cfg := config{}

	fmt.Printf("Loading configuration from: %v\n", configFile)

	readB, err := os.ReadFile(configFile)

	Assert(t, err == nil, fmt.Sprintf("Error reading file: %v, err: %v", configFile, err))

	reader := bytes.NewReader(readB)
	err = json.NewDecoder(reader).Decode(&cfg)
	Assert(t, err == nil, fmt.Sprintf("Error decoding file: %v, err: %v", configFile, err))

	fmt.Printf("config: %+v\n", cfg)

	if isEmpty {
		// verify the config is removed
		Assert(t, cfg.Features.Cdi == false, "CDI is enabled in the config file")
		Assert(t, len(cfg.Runtimes) == 0, "Number of runtimes in config is not 0")
		Assert(t, cfg.DefaultRuntime == "", fmt.Sprintf("default runtime should not be set to %v", cfg.DefaultRuntime))

	} else {
		Assert(t, cfg.Features.Cdi == true, "CDI is not enabled in the config file")
		Assert(t, len(cfg.Runtimes) == 1, "Number of runtimes in config is not 1")

		rtime, exists := cfg.Runtimes["amd"]
		Assert(t, exists == true, "amd runtime doesn't exist in the config file")
		Assert(t, rtime.Path == amdRuntimePath, fmt.Sprintf("amd runtime path isn't set to %v", amdRuntimePath))

		if isDefault {
			Assert(t, cfg.DefaultRuntime == amdRuntimeName, fmt.Sprintf("default runtime not set to %v", amdRuntimeName))
		} else {
			Assert(t, cfg.DefaultRuntime == "", fmt.Sprintf("default runtime should not be set to %v", cfg.DefaultRuntime))
		}
	}
}

func Assert(t *testing.T, b bool, errString string) {
	if !b {
		t.Errorf(errString)
	}
}

func cleanUp() {
	fmt.Printf("Deleting file: %v\n", configFile)
	os.Remove(configFile)
}

func TestConfigureRunTimeAddRemove(t *testing.T) {
	fmt.Printf("amd-ctk path: %v\n", cliPath)
	setup(t)
	cfgPathArg := "--config-path=" + configFile
	// add amd to runtimes
	out, outErr, err := runCLI("runtime", "configure", "--runtime=docker", cfgPathArg)

	Assert(t, outErr == "", fmt.Sprintf("amd-ctk runtime configure returned err: %v", outErr))
	Assert(t, err == nil, fmt.Sprintf("Error running amd-ctk err: %v", err))

	fmt.Println("output: ", out)
	verifyConfigFile(t, false, false)

	// remove amd from runtimes
	out, outErr, err = runCLI("runtime", "configure", "--runtime=docker", cfgPathArg, "--remove")

	Assert(t, outErr == "", fmt.Sprintf("amd-ctk runtime configure remove returned err: %v", outErr))
	Assert(t, err == nil, fmt.Sprintf("Error running amd-ctk err: %v", err))

	fmt.Println("output: ", out)
	verifyConfigFile(t, false, true)
	cleanUp()
}

func TestConfigureRunTimeDefault(t *testing.T) {
	fmt.Printf("amd-ctk path: %v\n", cliPath)
	setup(t)
	cfgPathArg := "--config-path=" + configFile
	// add amd to runtimes as default
	out, outErr, err := runCLI("runtime", "configure", "--runtime=docker", cfgPathArg, "--set-as-default")

	Assert(t, outErr == "", fmt.Sprintf("amd-ctk runtime configure as default returned err: %v", outErr))
	Assert(t, err == nil, fmt.Sprintf("Error running amd-ctk err: %v", err))

	fmt.Println("output: ", out)
	verifyConfigFile(t, true, false)

	// unset as default
	out, outErr, err = runCLI("runtime", "configure", "--runtime=docker", cfgPathArg, "--unset-as-default")

	Assert(t, outErr == "", fmt.Sprintf("amd-ctk runtime configure unset default returned err: %v", outErr))
	Assert(t, err == nil, fmt.Sprintf("Error running amd-ctk err: %v", err))

	fmt.Println("output: ", out)
	verifyConfigFile(t, false, false)

	// add it back
	out, outErr, err = runCLI("runtime", "configure", "--runtime=docker", cfgPathArg, "--set-as-default")

	Assert(t, outErr == "", fmt.Sprintf("amd-ctk runtime configure as default returned err: %v", outErr))
	Assert(t, err == nil, fmt.Sprintf("Error running amd-ctk err: %v", err))

	fmt.Println("output: ", out)
	verifyConfigFile(t, true, false)

	// use remove flag and make sure default gets deleted too
	out, outErr, err = runCLI("runtime", "configure", "--runtime=docker", cfgPathArg, "--remove")

	Assert(t, outErr == "", fmt.Sprintf("amd-ctk runtime configure unset default returned err: %v", outErr))
	Assert(t, err == nil, fmt.Sprintf("Error running amd-ctk err: %v", err))

	fmt.Println("output: ", out)
	verifyConfigFile(t, false, true)
	cleanUp()

}

func TestConfigureRuntimeMultiFlags(t *testing.T) {
	fmt.Printf("amd-ctk path: %v\n", cliPath)
	setup(t)
	cfgPathArg := "--config-path=" + configFile
	// add amd to runtimes as default along with remove flag
	out, outErr, err := runCLI("runtime", "configure", "--runtime=docker", cfgPathArg, "--remove", "--set-as-default")

	Assert(t, outErr == "", fmt.Sprintf("amd-ctk runtime configure as default returned err: %v", outErr))

	// shoudl error
	Assert(t, err != nil, "err shouldn't be nil")

	// match the error message
	Assert(t, strings.TrimSpace(out) == removeAndDefErrMsg, fmt.Sprintf("stdout: %v should have been '%v'", out, removeAndDefErrMsg))

	// use set default and unset default flags at the same time
	out, outErr, err = runCLI("runtime", "configure", "--runtime=docker", cfgPathArg, "--unset-as-default", "--set-as-default")

	Assert(t, outErr == "", fmt.Sprintf("amd-ctk runtime configure as default returned err: %v", outErr))

	// shoudl error
	Assert(t, err != nil, "err shouldn't be nil")

	// match the error message
	Assert(t, strings.TrimSpace(out) == setUnsetDefErrMsg, fmt.Sprintf("stdout: %v should have been '%v'", out, setUnsetDefErrMsg))
	cleanUp()
}
