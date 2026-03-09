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

package symlinks

import (
	"fmt"
	"os"

	"github.com/ROCm/container-toolkit/internal/logger"
)

// Resolve resolves a symlink to its target
// Returns os.ErrNotExist if the link doesn't exist
func Resolve(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return "", fmt.Errorf("%s is not a symlink", path)
	}

	target, err := os.Readlink(path)
	if err != nil {
		return "", fmt.Errorf("failed to read symlink: %v", err)
	}

	return target, nil
}

// ForceCreate creates a symlink, removing any existing file/link at that location
func ForceCreate(target, link string) error {
	// Remove existing file or symlink
	if _, err := os.Lstat(link); err == nil {
		logger.Log.Printf("Removing existing file/link at %s", link)
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("failed to remove existing link: %v", err)
		}
	}

	// Create the symlink
	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}

	return nil
}

// Mark creates a .created marker file to track created symlinks
func Mark(path string) error {
	markerPath := path + ".created"
	file, err := os.Create(markerPath)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

// IsMarked checks if a path has been marked as created
func IsMarked(path string) bool {
	markerPath := path + ".created"
	_, err := os.Stat(markerPath)
	return err == nil
}
