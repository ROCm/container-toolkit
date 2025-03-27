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

package amdgpu

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ROCm/container-toolkit/internal/logger"
)

// AMDGPU collects device information of GPU
type AMDGPU struct {
	Path     string
	Major    int64
	Minor    int64
	FileMode os.FileMode
	Gid      uint32
	Allow    bool
	DevType  string
	Access   string
}

// GetAMDGPUs returns the lists of all the GPU devices on the system.
// All devices under the same "pci:amdgpu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*"
// directory are in a single list.
// There are as many such lists as the number of gpu directories under "pci:amdgpu/".
func GetAMDGPUs() ([][]string, error) {
	if _, err := os.Stat("/sys/module/amdgpu/drivers/"); err != nil {
		logger.Log.Printf("amdgpu driver unavailable: %s", err)
		return nil, err
	}

	pciDevs, err := filepath.Glob("/sys/module/amdgpu/drivers/pci:amdgpu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*")
	if err != nil {
		logger.Log.Printf("Failed to find amdgpu driver directories: %s", err)
		return nil, err
	}
	// TODO: Is sort required?

	devs := [][]string{}
	for _, path := range pciDevs {
		drms, err := filepath.Glob(path + "/drm/*")
		if err != nil {
			logger.Log.Printf("Failed to find amdgpu driver drm directories: %s", err)
			return nil, err
		}

		drmDevs := []string{}
		for _, drm := range drms {
			dev := filepath.Base(drm)
			if dev[0:4] == "card" || dev[0:7] == "renderD" {
				drmDevs = append(drmDevs, "/dev/dri/"+dev)
			}
		}
		devs = append(devs, drmDevs)
	}

	return devs, nil
}

// GetAMDGPU returns device information for the given GPU
func GetAMDGPU(dev string) (AMDGPU, error) {
	getStat := func(dev, format string) (string, error) {
		out, err := exec.Command("stat", "-c", format, dev).Output()
		if err != nil {
			logger.Log.Printf("stat failed for %v, Error: %v", dev, err)
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	}

	convStat := func(dev, format string, base, width int) int64 {
		out, err := getStat(dev, format)
		if err != nil {
			return 0
		}

		var ret int64
		if base == 8 {
			ret, err = strconv.ParseInt("020"+out, base, width)
		} else {
			ret, err = strconv.ParseInt(out, base, width)
		}
		if err != nil {
			logger.Log.Printf("Failed to convert string %v to (%v, %v), Error: %v", ret, base, width, err)
			return 0
		}

		return ret
	}

	gpu := AMDGPU{
		Path:     dev,
		Major:    convStat(dev, "%t", 16, 64),
		Minor:    convStat(dev, "%T", 16, 64),
		FileMode: os.FileMode(convStat(dev, "%a", 8, 64)),
		Allow:    true,
		DevType:  "c",
		Access:   "rwm",
	}

	gid := convStat(dev, "%g", 10, 32)
	gpu.Gid = uint32(gid)

	return gpu, nil
}

/*
func getHexStat(dev, format string) int64 {
	out, err := exec.Command("stat", "-c", format, dev).Output()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	value, err := strconv.ParseInt(strings.TrimSpace(string(out)), 16, 64)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	return value
}

func getOctalStat(dev, format string) int64 {
	out, err := exec.Command("stat", "-c", format, dev).Output()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	value, err := strconv.ParseInt("020"+strings.TrimSpace(string(out)), 8, 64)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	return value
}

func getStat(dev, format string) string {
	out, err := exec.Command("stat", "-c", format, dev).Output()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(out))
}
*/
