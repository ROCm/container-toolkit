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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ROCm/container-toolkit/internal/logger"
	"github.com/moby/sys/symlink"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli/v2"
)

type command struct{}

type config struct {
	links         []string
	containerSpec string
}

// NewCommand constructs the create-symlinks hook command
func NewCommand() *cli.Command {
	c := command{}
	return c.build()
}

func (m command) build() *cli.Command {
	cfg := config{}

	return &cli.Command{
		Name:  "create-symlinks",
		Usage: "Create symlinks in the container for legacy ROCm paths",
		Action: func(_ context.Context, cmd *cli.Command) error {
			return m.run(cmd, &cfg)
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "link",
				Usage:       "Symlink specification: target::link. Example: /opt/rocm/lib::/opt/rocm-5.7.0/lib",
				Destination: &cfg.links,
			},
			&cli.StringFlag{
				Name:        "container-spec",
				Usage:       "Path to OCI container spec (for testing)",
				Destination: &cfg.containerSpec,
				Hidden:      true,
			},
		},
	}
}

func (m command) run(_ *cli.Command, cfg *config) error {
	containerRoot, err := m.getContainerRoot(cfg.containerSpec)
	if err != nil {
		return fmt.Errorf("failed to determine container root: %w", err)
	}

	created := make(map[string]bool)
	for _, l := range cfg.links {
		if created[l] {
			logger.Log.Printf("Link %v already processed", l)
			continue
		}
		parts := strings.Split(l, "::")
		if len(parts) != 2 {
			return fmt.Errorf("invalid symlink specification %v (expected target::link)", l)
		}

		if err := m.createLink(containerRoot, parts[0], parts[1]); err != nil {
			return fmt.Errorf("failed to create link %v: %w", parts, err)
		}
		created[l] = true
	}
	return nil
}

// getContainerRoot determines the container root from the OCI spec
func (m command) getContainerRoot(specPath string) (string, error) {
	if specPath == "" || specPath == "-" {
		specPath = "/dev/stdin"
	}

	file, err := os.Open(specPath)
	if err != nil {
		return "", fmt.Errorf("failed to open spec: %w", err)
	}
	defer file.Close()

	var spec specs.Spec
	if err := json.NewDecoder(file).Decode(&spec); err != nil {
		return "", fmt.Errorf("failed to decode spec: %w", err)
	}

	if spec.Root == nil {
		return "", fmt.Errorf("spec.Root is nil")
	}

	return spec.Root.Path, nil
}

// createLink creates a symbolic link in the container root
func (m command) createLink(containerRoot, targetPath, linkPath string) error {
	fullLinkPath := filepath.Join(containerRoot, linkPath)

	// Check if link already exists with correct target
	exists, err := linkExists(targetPath, fullLinkPath)
	if err != nil {
		return fmt.Errorf("failed to check link existence: %w", err)
	}
	if exists {
		logger.Log.Printf("Link %s already exists with correct target", fullLinkPath)
		return nil
	}

	// Resolve parent directory within container root
	resolvedParent, err := symlink.FollowSymlinkInScope(filepath.Dir(fullLinkPath), containerRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve link parent: %w", err)
	}
	resolvedLinkPath := filepath.Join(resolvedParent, filepath.Base(fullLinkPath))

	logger.Log.Printf("Creating symlink: %s -> %s", resolvedLinkPath, targetPath)

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(resolvedLinkPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Remove existing file/link if present
	if err := os.Remove(resolvedLinkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing path: %w", err)
	}

	// Create symlink
	if err := os.Symlink(targetPath, resolvedLinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// linkExists checks if a symlink exists and points to the expected target
func linkExists(target, link string) (bool, error) {
	currentTarget, err := os.Readlink(link)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		// Not a symlink or other error
		return false, nil
	}
	return currentTarget == target, nil
}
