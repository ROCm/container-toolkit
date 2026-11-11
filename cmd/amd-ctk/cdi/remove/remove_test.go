package remove

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestRemoveCommand(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		create      bool
		asDir       bool
		wantErr     bool
		wantRemoved bool
	}{
		{name: "AMD spec is removed", content: `{"kind": "amd.com/gpu"}`, create: true, wantRemoved: true},
		{name: "missing spec succeeds"},
		{name: "foreign spec is kept", content: `{"kind": "vendor.example/device"}`, create: true},
		{name: "malformed spec is kept", content: `{bad`, create: true},
		{name: "unreadable path fails", asDir: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			other := filepath.Join(dir, "acme.json")
			assert.NoError(t, os.WriteFile(other, []byte(`{"kind": "vendor.example/device"}`), 0644))
			spec := filepath.Join(dir, "amd.json")
			if tt.create {
				assert.NoError(t, os.WriteFile(spec, []byte(tt.content), 0644))
			}
			if tt.asDir {
				assert.NoError(t, os.Mkdir(spec, 0755))
			}

			app := &cli.App{Commands: []*cli.Command{AddNewCommand()}}
			err := app.Run([]string{"amd-ctk", "remove", "--path", spec})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			_, statErr := os.Stat(spec)
			if tt.wantRemoved {
				assert.True(t, os.IsNotExist(statErr), "spec should be removed")
			} else if tt.create || tt.asDir {
				assert.NoError(t, statErr, "spec should be kept")
			}
			assert.FileExists(t, other, "third-party spec must survive")
		})
	}
}
