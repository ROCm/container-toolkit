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
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ROCm/container-toolkit/internal/logger"
)

// FileSystem interface for mocking filesystem operations
type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	Glob(pattern string) ([]string, error)
	ReadFile(name string) ([]byte, error)
	GetDeviceStat(dev string, format string) (string, error)
}

// DefaultFS implements FileSystem using actual filesystem operations
type DefaultFS struct{}

func (fs *DefaultFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fs *DefaultFS) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func (fs *DefaultFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (fs *DefaultFS) GetDeviceStat(dev string, format string) (string, error) {
	out, err := exec.Command("stat", "-c", format, dev).Output()
	if err != nil {
		logger.Log.Printf("stat failed for %v, Error: %v", dev, err)
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

var defaultFS FileSystem = &DefaultFS{}

// AMDGPU collects device information of GPU
type AMDGPU struct {
	Path     string
	Major    int64
	Minor    int64
	FileMode os.FileMode
	Gid      uint32
	Uid      uint32
	Allow    bool
	DevType  string
	Access   string
}

// GetAMDGPUs returns the lists of all the GPU devices on the system.
// All devices under the same "pci:amdgpu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*"
// directory are in a single list.
// There are as many such lists as the number of gpu directories under "pci:amdgpu/".
func GetAMDGPUs() ([][]string, error) {
	return GetAMDGPUsWithFS(defaultFS)
}

// GetAMDGPUsWithFS is the internal implementation that takes a FileSystem interface.
// This split allows for better testability by enabling mock filesystem operations
// in unit tests, while keeping the public API simple with GetAMDGPUs.
func GetAMDGPUsWithFS(fs FileSystem) ([][]string, error) {
	if _, err := fs.Stat("/sys/module/amdgpu/drivers/"); err != nil {
		logger.Log.Printf("amdgpu driver unavailable: %s", err)
		return nil, err
	}

	renderDevIds := GetDevIdsFromTopology(fs)

	// Map to store devices by unique_id to maintain grouping
	uniqueIdDevices := make(map[string][][]string)
	var uniqueIds []string // To maintain order

	// Process PCI devices
	pciDevs, err := fs.Glob("/sys/module/amdgpu/drivers/pci:amdgpu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*")
	if err != nil {
		logger.Log.Printf("Failed to find amdgpu driver directories: %s", err)
		return nil, err
	}

	// Process platform devices for partitions
	platformDevs, _ := fs.Glob("/sys/devices/platform/amdgpu_xcp_*")

	// Combine both PCI and platform devices
	allDevs := append(pciDevs, platformDevs...)

	// Process all devices using the same logic
	for _, path := range allDevs {
		drms, err := fs.Glob(path + "/drm/*")
		if err != nil {
			logger.Log.Printf("Failed to find amdgpu driver drm directories: %s", err)
			return nil, err
		}

		drmDevs := []string{}
		renderMinor := 0
		for _, drm := range drms {
			dev := filepath.Base(drm)
			if len(dev) >= 4 && dev[0:4] == "card" || len(dev) >= 7 && dev[0:7] == "renderD" {
				drmDevs = append(drmDevs, "/dev/dri/"+dev)
				if len(dev) >= 7 && dev[0:7] == "renderD" {
					renderMinor, _ = strconv.Atoi(dev[7:])
				}
			}
		}

		if len(drmDevs) > 0 && renderMinor > 0 {
			if devID, exists := renderDevIds[renderMinor]; exists {
				if _, exists := uniqueIdDevices[devID]; !exists {
					uniqueIds = append(uniqueIds, devID)
				}
				uniqueIdDevices[devID] = append(uniqueIdDevices[devID], drmDevs)
			}
		}
	}

	// Sort devices within each unique_id group by render minor number
	for _, devID := range uniqueIds {
		sort.Slice(uniqueIdDevices[devID], func(i, j int) bool {
			getRenderID := func(devs []string) int {
				for _, dev := range devs {
					baseDev := filepath.Base(dev)
					if len(baseDev) >= 7 && strings.HasPrefix(baseDev, "renderD") {
						id, _ := strconv.Atoi(strings.TrimPrefix(baseDev, "renderD"))
						return id
					}
				}
				return 0
			}
			return getRenderID(uniqueIdDevices[devID][i]) < getRenderID(uniqueIdDevices[devID][j])
		})
	}

	// Combine all devices maintaining the unique_id order
	var devs [][]string
	for _, devID := range uniqueIds {
		devs = append(devs, uniqueIdDevices[devID]...)
	}

	return devs, nil
}

// GetAMDGPU returns device information for the given GPU
func GetAMDGPU(dev string) (AMDGPU, error) {
	return GetAMDGPUWithFS(defaultFS, dev)
}

func GetAMDGPUWithFS(fs FileSystem, dev string) (AMDGPU, error) {
	getStat := func(dev, format string) (string, error) {
		return fs.GetDeviceStat(dev, format)
	}

	convStat := func(dev, format string, base, width int) int64 {
		out, err := getStat(dev, format)
		if err != nil {
			return 0
		}

		var ret int64
		if base == 8 {
			ret, err = strconv.ParseInt("0"+out, base, width)
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

	uid := 0
	gpu.Uid = uint32(uid)

	return gpu, nil
}

var topoUniqueIdRe = regexp.MustCompile(`unique_id\s(\d+)`)
var renderMinorRe = regexp.MustCompile(`drm_render_minor\s(\d+)`)

// GetDevIdsFromTopology returns a map of render minor numbers to unique_ids
func GetDevIdsFromTopology(fs FileSystem, topoRootParam ...string) map[int]string {
	topoRoot := "/sys/class/kfd/kfd"
	if len(topoRootParam) == 1 {
		topoRoot = topoRootParam[0]
	}

	renderDevIds := make(map[int]string)
	nodeFiles, err := fs.Glob(topoRoot + "/topology/nodes/*/properties")
	if err != nil {
		logger.Log.Printf("glob error: %s", err)
		return renderDevIds
	}

	for _, nodeFile := range nodeFiles {
		logger.Log.Printf("Parsing %s", nodeFile)
		renderMinor, err := ParseTopologyProperties(fs, nodeFile, renderMinorRe)
		if err != nil {
			logger.Log.Printf("Error parsing render minor: %v", err)
			continue
		}

		if renderMinor <= 0 || renderMinor > math.MaxInt32 {
			continue
		}

		devID, err := ParseTopologyPropertiesString(fs, nodeFile, topoUniqueIdRe)
		if err != nil {
			logger.Log.Printf("Error parsing unique_id: %v", err)
			continue
		}

		renderDevIds[int(renderMinor)] = devID
	}

	return renderDevIds
}

// ParseTopologyProperties parses for a property value in kfd topology file as int64
// The format is usually one entry per line <name> <value>.
func ParseTopologyProperties(fs FileSystem, path string, re *regexp.Regexp) (int64, error) {
	content, err := fs.ReadFile(path)
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if matches != nil {
			return strconv.ParseInt(matches[1], 0, 64)
		}
	}

	return 0, fmt.Errorf("property not found in %s", path)
}

// ParseTopologyPropertiesString parses for a property value in kfd topology file as string
// The format is usually one entry per line <name> <value>.
func ParseTopologyPropertiesString(fs FileSystem, path string, re *regexp.Regexp) (string, error) {
	content, err := fs.ReadFile(path)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if matches != nil {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("property not found in %s", path)
}
