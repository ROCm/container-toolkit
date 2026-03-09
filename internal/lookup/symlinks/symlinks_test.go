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
	"os"
	"path/filepath"
	"testing"

	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	logger.Init(true)
	tmpDir := t.TempDir()

	target := "target-file"
	linkPath := filepath.Join(tmpDir, "testlink")
	require.NoError(t, os.Symlink(target, linkPath))

	resolved, err := Resolve(linkPath)
	require.NoError(t, err)
	assert.Equal(t, target, resolved)

	// Test non-symlink
	regularFile := filepath.Join(tmpDir, "regular")
	require.NoError(t, os.WriteFile(regularFile, []byte("test"), 0644))
	_, err = Resolve(regularFile)
	assert.Error(t, err)

	// Test non-existent
	_, err = Resolve(filepath.Join(tmpDir, "nonexistent"))
	assert.Error(t, err)
}

func TestForceCreate(t *testing.T) {
	logger.Init(true)
	tmpDir := t.TempDir()

	target := "new-target"
	linkPath := filepath.Join(tmpDir, "link")

	// Create initial symlink
	require.NoError(t, ForceCreate("old-target", linkPath))
	resolved, err := Resolve(linkPath)
	require.NoError(t, err)
	assert.Equal(t, "old-target", resolved)

	// Force create with new target
	require.NoError(t, ForceCreate(target, linkPath))
	resolved, err = Resolve(linkPath)
	require.NoError(t, err)
	assert.Equal(t, target, resolved)

	// Force create over regular file
	regularFile := filepath.Join(tmpDir, "regular")
	require.NoError(t, os.WriteFile(regularFile, []byte("test"), 0644))
	require.NoError(t, ForceCreate(target, regularFile))
	resolved, err = Resolve(regularFile)
	require.NoError(t, err)
	assert.Equal(t, target, resolved)
}

func TestMarkAndIsMarked(t *testing.T) {
	logger.Init(true)
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "somefile")

	assert.False(t, IsMarked(path))

	require.NoError(t, Mark(path))
	assert.True(t, IsMarked(path))

	// Verify marker file exists
	markerPath := path + ".created"
	_, err := os.Stat(markerPath)
	require.NoError(t, err)
}
