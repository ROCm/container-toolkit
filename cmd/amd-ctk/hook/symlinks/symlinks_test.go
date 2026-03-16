package symlinks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) {
	logger.Init(true)
}

func TestLinkExists(t *testing.T) {
	setup(t)
	tmpDir := t.TempDir()

	// Create a symlink
	target := "target-file"
	link := filepath.Join(tmpDir, "test-link")
	require.NoError(t, os.Symlink(target, link))

	// Test existing link with correct target
	exists, err := linkExists(target, link)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test existing link with wrong target
	exists, err = linkExists("different-target", link)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test non-existent link
	exists, err = linkExists("foo", filepath.Join(tmpDir, "nonexistent"))
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestCreateLink(t *testing.T) {
	setup(t)
	cmd := command{}

	tests := []struct {
		name          string
		target        string
		linkPath      string
		setup         func(t *testing.T, root string)
		wantErr       bool
		validateLink  func(t *testing.T, root, target, linkPath string)
	}{
		{
			name:     "simple_relative_link",
			target:   "librocm.so.5",
			linkPath: "/opt/rocm/lib/librocm.so",
			validateLink: func(t *testing.T, root, target, linkPath string) {
				fullPath := filepath.Join(root, linkPath)
				resolvedTarget, err := os.Readlink(fullPath)
				require.NoError(t, err)
				assert.Equal(t, target, resolvedTarget)
			},
		},
		{
			name:     "absolute_target",
			target:   "/opt/rocm-5.7.0/lib/libhip.so",
			linkPath: "/opt/rocm/lib/libhip.so",
			validateLink: func(t *testing.T, root, target, linkPath string) {
				fullPath := filepath.Join(root, linkPath)
				resolvedTarget, err := os.Readlink(fullPath)
				require.NoError(t, err)
				assert.Equal(t, target, resolvedTarget)
			},
		},
		{
			name:     "nested_directory_creation",
			target:   "../lib/librocm.so",
			linkPath: "/opt/rocm-5.7.0/compat/lib/librocm.so",
			validateLink: func(t *testing.T, root, target, linkPath string) {
				fullPath := filepath.Join(root, linkPath)
				_, err := os.Stat(fullPath)
				require.NoError(t, err)
			},
		},
		{
			name:     "overwrites_existing_link",
			target:   "new-target",
			linkPath: "/test/link",
			setup: func(t *testing.T, root string) {
				linkPath := filepath.Join(root, "/test/link")
				require.NoError(t, os.MkdirAll(filepath.Dir(linkPath), 0755))
				require.NoError(t, os.Symlink("old-target", linkPath))
			},
			validateLink: func(t *testing.T, root, target, linkPath string) {
				fullPath := filepath.Join(root, linkPath)
				resolvedTarget, err := os.Readlink(fullPath)
				require.NoError(t, err)
				assert.Equal(t, target, resolvedTarget)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerRoot := t.TempDir()

			if tt.setup != nil {
				tt.setup(t, containerRoot)
			}

			err := cmd.createLink(containerRoot, tt.target, tt.linkPath)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.validateLink != nil {
				tt.validateLink(t, containerRoot, tt.target, tt.linkPath)
			}
		})
	}
}

func TestGetContainerRoot(t *testing.T) {
	setup(t)
	cmd := command{}

	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "config.json")

	// Create a minimal OCI spec
	spec := `{
		"ociVersion": "1.0.0",
		"root": {
			"path": "/container/root"
		}
	}`

	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))

	root, err := cmd.getContainerRoot(specPath)
	require.NoError(t, err)
	assert.Equal(t, "/container/root", root)
}

func TestCreateLinkIdempotent(t *testing.T) {
	setup(t)
	cmd := command{}

	containerRoot := t.TempDir()
	target := "librocm.so.5"
	linkPath := "/opt/rocm/lib/librocm.so"

	// Create link first time
	err := cmd.createLink(containerRoot, target, linkPath)
	require.NoError(t, err)

	// Create same link second time (should succeed)
	err = cmd.createLink(containerRoot, target, linkPath)
	require.NoError(t, err)

	// Verify link still points to correct target
	fullPath := filepath.Join(containerRoot, linkPath)
	resolvedTarget, err := os.Readlink(fullPath)
	require.NoError(t, err)
	assert.Equal(t, target, resolvedTarget)
}
