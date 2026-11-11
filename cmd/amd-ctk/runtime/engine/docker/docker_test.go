package docker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	amdSpec     = `{"cdiVersion": "0.5.0", "kind": "amd.com/gpu", "devices": []}`
	foreignSpec = `{"cdiVersion": "0.5.0", "kind": "vendor.example/device", "devices": []}`
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func TestRemoveAMDSpec(t *testing.T) {
	tests := []struct {
		name        string
		amdJSON     string
		createAMD   bool
		wantRemoved bool
	}{
		{name: "AMD spec is removed", amdJSON: amdSpec, createAMD: true, wantRemoved: true},
		{name: "foreign spec named amd.json is kept", amdJSON: foreignSpec, createAMD: true},
		{name: "malformed amd.json is kept", amdJSON: "not json", createAMD: true},
		{name: "missing amd.json is not an error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			other := filepath.Join(dir, "acme.json")
			writeFile(t, other, foreignSpec)
			amd := filepath.Join(dir, "amd.json")
			if tt.createAMD {
				writeFile(t, amd, tt.amdJSON)
			}

			assert.NoError(t, removeAMDSpec(dir))

			assert.FileExists(t, other, "third-party spec must survive")
			if tt.createAMD {
				_, err := os.Stat(amd)
				assert.Equal(t, tt.wantRemoved, os.IsNotExist(err), "amd.json removed")
			}
		})
	}
}

func TestRemoveAMDSpec_MissingDir(t *testing.T) {
	assert.NoError(t, removeAMDSpec(filepath.Join(t.TempDir(), "absent")))
}

func TestRemoveRuntime(t *testing.T) {
	tests := []struct {
		name            string
		config          string
		want            string
		wantErr         bool
		wantDoNotUpdate bool
	}{
		{
			name: "only AMD settings are removed",
			config: `{
				"default-runtime": "amd",
				"runtimes": {"amd": {"path": "amd-container-runtime"}, "other": {"path": "other-runtime"}},
				"features": {"cdi": true, "containerd-snapshotter": true},
				"log-level": "debug"
			}`,
			want: `{
				"runtimes": {"other": {"path": "other-runtime"}},
				"features": {"containerd-snapshotter": true},
				"log-level": "debug"
			}`,
		},
		{
			name:   "empty maps are dropped",
			config: `{"default-runtime": "amd", "runtimes": {"amd": {}}, "features": {"cdi": true}}`,
			want:   `{}`,
		},
		{
			name:   "foreign default runtime is kept",
			config: `{"default-runtime": "other", "runtimes": {"amd": {}, "other": {}}}`,
			want:   `{"default-runtime": "other", "runtimes": {"other": {}}}`,
		},
		{
			name:   "cdi feature is removed without a runtimes key",
			config: `{"features": {"cdi": true}}`,
			want:   `{}`,
		},
		{
			name:   "default runtime is removed without a runtimes key",
			config: `{"default-runtime": "amd"}`,
			want:   `{}`,
		},
		{
			name:            "nothing to remove",
			config:          `{"log-level": "debug"}`,
			want:            `{"log-level": "debug"}`,
			wantDoNotUpdate: true,
		},
		{
			name:            "non-string default runtime is left alone",
			config:          `{"default-runtime": 1}`,
			want:            `{"default-runtime": 1}`,
			wantDoNotUpdate: true,
		},
		{
			name:            "malformed runtimes is an error",
			config:          `{"runtimes": "x"}`,
			wantErr:         true,
			wantDoNotUpdate: true,
		},
		{
			name:            "malformed features is an error",
			config:          `{"features": ["cdi"]}`,
			wantErr:         true,
			wantDoNotUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cdiSpecDir = t.TempDir()
			t.Cleanup(func() { cdiSpecDir = defaultCDISpecPath })

			cfg := dockerConfig{}
			if err := json.Unmarshal([]byte(tt.config), &cfg); err != nil {
				t.Fatalf("decoding test config: %v", err)
			}

			err, doNotUpdate := cfg.RemoveRuntime("amd")

			assert.Equal(t, tt.wantDoNotUpdate, doNotUpdate, "doNotUpdate")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			want := dockerConfig{}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatalf("decoding expected config: %v", err)
			}
			assert.Equal(t, want, cfg)
		})
	}
}

func TestRemoveRuntime_KeepsThirdPartySpecs(t *testing.T) {
	cdiSpecDir = t.TempDir()
	t.Cleanup(func() { cdiSpecDir = defaultCDISpecPath })
	other := filepath.Join(cdiSpecDir, "acme.json")
	amd := filepath.Join(cdiSpecDir, "amd.json")
	writeFile(t, other, foreignSpec)
	writeFile(t, amd, amdSpec)

	cfg := dockerConfig{}
	err, _ := cfg.RemoveRuntime("amd")

	assert.NoError(t, err)
	assert.FileExists(t, other, "third-party spec must survive")
	assert.NoFileExists(t, amd, "AMD spec must be removed")
}
