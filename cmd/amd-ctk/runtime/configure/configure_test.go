package configure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestRemoveCDISpec(t *testing.T) {
	tests := []struct {
		name        string
		args        func(systemConfig, dir string) []string
		config      string
		wantErr     bool
		wantRemoved bool
	}{
		{
			name:        "system config removes the spec",
			args:        func(sys, _ string) []string { return []string{"--remove", "--config-path=" + sys} },
			config:      `{"runtimes": {"amd": {}}}`,
			wantRemoved: true,
		},
		{
			name: "system config spelled differently removes the spec",
			args: func(_, dir string) []string {
				return []string{"--remove", "--config-path=" + dir + "/docker/../docker//daemon.json"}
			},
			config:      `{"runtimes": {"amd": {}}}`,
			wantRemoved: true,
		},
		{
			name:        "malformed system config still removes the spec",
			args:        func(sys, _ string) []string { return []string{"--remove", "--config-path=" + sys} },
			config:      `{bad`,
			wantErr:     true,
			wantRemoved: true,
		},
		{
			name: "custom config keeps the spec",
			args: func(_, dir string) []string {
				return []string{"--remove", "--config-path=" + filepath.Join(dir, "custom.json")}
			},
			config: `{"runtimes": {"amd": {}}}`,
		},
		{
			name: "rejected flags keep the spec",
			args: func(sys, _ string) []string {
				return []string{"--remove", "--set-as-default", "--config-path=" + sys}
			},
			config:  `{"runtimes": {"amd": {}}}`,
			wantErr: true,
		},
		{
			name:   "configure without remove keeps the spec",
			args:   func(sys, _ string) []string { return []string{"--config-path=" + sys} },
			config: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			specDir := filepath.Join(dir, "cdi")
			systemConfig := filepath.Join(dir, "docker", "daemon.json")
			assert.NoError(t, os.MkdirAll(specDir, 0755))
			assert.NoError(t, os.MkdirAll(filepath.Dir(systemConfig), 0755))

			amdSpec := filepath.Join(specDir, "amd.json")
			otherSpec := filepath.Join(specDir, "acme.json")
			assert.NoError(t, os.WriteFile(amdSpec, []byte(`{"kind": "amd.com/gpu"}`), 0644))
			assert.NoError(t, os.WriteFile(otherSpec, []byte(`{"kind": "vendor.example/device"}`), 0644))
			assert.NoError(t, os.WriteFile(systemConfig, []byte(tt.config), 0644))

			savedSpecDir, savedConfig := cdiSpecDir, systemDockerConfigPath
			cdiSpecDir, systemDockerConfigPath = specDir, systemConfig
			t.Cleanup(func() { cdiSpecDir, systemDockerConfigPath = savedSpecDir, savedConfig })

			app := &cli.App{Commands: []*cli.Command{AddNewCommand()}}
			err := app.Run(append([]string{"amd-ctk", "configure"}, tt.args(systemConfig, dir)...))

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			_, statErr := os.Stat(amdSpec)
			assert.Equal(t, tt.wantRemoved, os.IsNotExist(statErr), "amd.json removed")
			assert.FileExists(t, otherSpec, "third-party spec must survive")
		})
	}
}
