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
	"github.com/ROCm/container-toolkit/internal/lookup/symlinks"
	"github.com/moby/sys/symlink"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli/v2"
)

type command struct{}

type config struct {
	links         []string
	containerSpec string
}

// AddNewCommand constructs the create-symlinks hook command
func AddNewCommand() *cli.Command {
	c := command{}
	return c.build()
}

// build creates the create-symlink command
func (m command) build() *cli.Command {
	cfg := config{}

	return &cli.Command{
		Name:  "create-symlinks",
		Usage: "A hook to create symlinks for legacy ROCm paths in the container",
		Action: func(_ context.Context, cmd *cli.Command) error {
			return m.run(cmd, &cfg)
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "link",
				Usage:       "Specify a link to create as target::link. If the link exists with different target, it is replaced.",
				Destination: &cfg.links,
			},
			&cli.StringFlag{
				Name:        "container-spec",
				Usage:       "Path to OCI container spec. If empty or '-', read from STDIN.",
				Destination: &cfg.containerSpec,
				Hidden:      true,
			},
		},
	}
}

func (m command) run(_ *cli.Command, cfg *config) error {
	containerRoot, err := m.getContainerRoot(cfg.containerSpec)
	if err != nil {
		return fmt.Errorf("failed to determine container root: %v", err)
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

		err := m.createLink(containerRoot, parts[0], parts[1])
		if err != nil {
			return fmt.Errorf("failed to create link %v: %w", parts, err)
		}
		created[l] = true
	}
	return nil
}

// getContainerRoot extracts container root from OCI spec
func (m command) getContainerRoot(specPath string) (string, error) {
	if specPath == "" || specPath == "-" {
		specPath = "/dev/stdin"
	}

	file, err := os.Open(specPath)
	if err != nil {
		return "", fmt.Errorf("failed to open spec: %v", err)
	}
	defer file.Close()

	var spec specs.Spec
	if err := json.NewDecoder(file).Decode(&spec); err != nil {
		return "", fmt.Errorf("failed to decode spec: %v", err)
	}

	if spec.Root == nil {
		return "", fmt.Errorf("spec missing root")
	}

	return spec.Root.Path, nil
}

// createLink creates a symbolic link in the container root
// Equivalent to: chroot <containerRoot> ln -sf <target> <link>
func (m command) createLink(containerRoot string, targetPath string, link string) error {
	linkPath := filepath.Join(containerRoot, link)

	exists, err := linkExists(targetPath, linkPath)
	if err != nil {
		return fmt.Errorf("failed to check if link exists: %w", err)
	}
	if exists {
		logger.Log.Printf("Link %s already exists with correct target", linkPath)
		return nil
	}

	// Resolve parent of the symlink in container root to prevent escape
	resolvedLinkParent, err := symlink.FollowSymlinkInScope(filepath.Dir(linkPath), containerRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve parent for link %v: %w", link, err)
	}
	resolvedLinkPath := filepath.Join(resolvedLinkParent, filepath.Base(linkPath))

	logger.Log.Printf("Creating symlink %v -> %v", resolvedLinkPath, targetPath)
	if err := os.MkdirAll(filepath.Dir(resolvedLinkPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	if err := symlinks.ForceCreate(targetPath, resolvedLinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}

	return nil
}

// linkExists checks if link exists and points to target
func linkExists(target string, link string) (bool, error) {
	currentTarget, err := symlinks.Resolve(link)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to resolve link %s: %w", link, err)
	}
	return currentTarget == target, nil
}
