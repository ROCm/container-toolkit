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
	"github.com/ROCm/container-toolkit/internal/lookup/symlinks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkExists(t *testing.T) {
	logger.Init(true)
	tmpDir := t.TempDir()
	require.NoError(t, makeFs(tmpDir,
		dirOrLink{path: "/a/b/c", target: "d"},
		dirOrLink{path: "/a/b/e", target: "/a/b/f"},
	))

	exists, err := linkExists("d", filepath.Join(tmpDir, "/a/b/c"))
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = linkExists("/a/b/f", filepath.Join(tmpDir, "/a/b/e"))
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = linkExists("different-target", filepath.Join(tmpDir, "/a/b/c"))
	require.NoError(t, err)
	require.False(t, exists)

	exists, err = linkExists("foo", filepath.Join(tmpDir, "/a/b/nonexistent"))
	require.NoError(t, err)
	require.False(t, exists)
}

func TestCreateLink(t *testing.T) {
	logger.Init(true)

	tests := []struct {
		name              string
		containerContents []dirOrLink
		target            string
		link              string
		wantErr           bool
		expectedTarget    string
	}{
		{
			name:              "simple_relative_link",
			containerContents: []dirOrLink{{path: "/opt/rocm/lib/"}},
			target:            "libamdhip64.so.5",
			link:              "/opt/rocm/lib/libamdhip64.so",
			expectedTarget:    "libamdhip64.so.5",
		},
		{
			name:              "absolute_target",
			containerContents: []dirOrLink{{path: "/opt/rocm-5.7.0/lib/"}},
			target:            "/opt/rocm-6.0.0/lib/libamdhip64.so.5",
			link:              "/opt/rocm-5.7.0/lib/libamdhip64.so",
			expectedTarget:    "/opt/rocm-6.0.0/lib/libamdhip64.so.5",
		},
		{
			name: "replace_existing_wrong_link",
			containerContents: []dirOrLink{
				{path: "/opt/rocm/lib/"},
				{path: "/opt/rocm/lib/libamdhip64.so", target: "wrong-target"},
			},
			target:         "libamdhip64.so.5",
			link:           "/opt/rocm/lib/libamdhip64.so",
			expectedTarget: "libamdhip64.so.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			containerRoot := filepath.Join(tmpDir, "container")
			require.NoError(t, makeFs(containerRoot, tt.containerContents...))

			cmd := command{}
			err := cmd.createLink(containerRoot, tt.target, tt.link)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			linkPath := filepath.Join(containerRoot, tt.link)
			target, err := symlinks.Resolve(linkPath)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTarget, target)
		})
	}
}

type dirOrLink struct {
	path   string
	target string
}

func makeFs(baseDir string, items ...dirOrLink) error {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return err
	}
	for _, item := range items {
		fullPath := filepath.Join(baseDir, item.path)
		if item.target == "" {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return err
			}
			if err := os.Symlink(item.target, fullPath); err != nil && !os.IsExist(err) {
				return err
			}
		}
	}
	return nil
}
