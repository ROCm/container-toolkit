package docker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigFile_StripsBOM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "daemon.json")
	assert.NoError(t, os.WriteFile(path, []byte("\xef\xbb\xbf{\"default-runtime\": \"amd\"}"), 0644))

	cfg, err := loadConfigFile(path)

	assert.NoError(t, err)
	if assert.NotNil(t, cfg) {
		assert.Equal(t, dockerConfig{"default-runtime": "amd"}, *cfg)
	}
}

func TestConfigRuntime(t *testing.T) {
	const amdRuntime = `{"path": "amd-container-runtime", "args": []}`
	tests := []struct {
		name      string
		config    string
		isDefault bool
		want      string
		wantErr   bool
	}{
		{
			name:   "other features and top-level keys are kept",
			config: `{"features": {"containerd-snapshotter": true}, "log-level": "debug"}`,
			want: `{
				"runtimes": {"amd": ` + amdRuntime + `},
				"features": {"containerd-snapshotter": true, "cdi": true},
				"log-level": "debug"
			}`,
		},
		{
			name:   "disabled cdi is enabled",
			config: `{"features": {"cdi": false}}`,
			want:   `{"runtimes": {"amd": ` + amdRuntime + `}, "features": {"cdi": true}}`,
		},
		{
			name:   "other runtimes are kept",
			config: `{"runtimes": {"nvidia": {"path": "nvidia-container-runtime"}}}`,
			want: `{
				"runtimes": {"nvidia": {"path": "nvidia-container-runtime"}, "amd": ` + amdRuntime + `},
				"features": {"cdi": true}
			}`,
		},
		{
			name:   "null sections are treated as absent",
			config: `{"runtimes": null, "features": null}`,
			want:   `{"runtimes": {"amd": ` + amdRuntime + `}, "features": {"cdi": true}}`,
		},
		{
			name:      "default runtime is set",
			config:    `{}`,
			isDefault: true,
			want: `{
				"default-runtime": "amd",
				"runtimes": {"amd": ` + amdRuntime + `},
				"features": {"cdi": true}
			}`,
		},
		{
			name:    "malformed features is an error",
			config:  `{"runtimes": {"nvidia": {}}, "features": ["cdi"]}`,
			wantErr: true,
		},
		{
			name:      "malformed runtimes is an error",
			config:    `{"runtimes": "x"}`,
			isDefault: true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := dockerConfig{}
			if err := json.Unmarshal([]byte(tt.config), &cfg); err != nil {
				t.Fatalf("decoding test config: %v", err)
			}

			err := cfg.ConfigRuntime("amd", "amd-container-runtime", tt.isDefault)

			want := tt.want
			if tt.wantErr {
				assert.Error(t, err)
				want = tt.config
			} else {
				assert.NoError(t, err)
			}
			assert.JSONEq(t, want, toJSON(t, cfg))
		})
	}
}

func toJSON(t *testing.T, cfg dockerConfig) string {
	t.Helper()
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("encoding config: %v", err)
	}
	return string(b)
}

func TestRemoveRuntime(t *testing.T) {
	tests := []struct {
		name            string
		config          string
		want            string
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
			name:   "malformed runtimes is skipped",
			config: `{"default-runtime": "amd", "runtimes": "x", "features": {"cdi": true}}`,
			want:   `{"runtimes": "x"}`,
		},
		{
			name:   "malformed features is skipped",
			config: `{"default-runtime": "amd", "runtimes": {"amd": {}}, "features": ["cdi"]}`,
			want:   `{"features": ["cdi"]}`,
		},
		{
			name:   "null sections are treated as absent",
			config: `{"default-runtime": "amd", "runtimes": null, "features": null}`,
			want:   `{"runtimes": null, "features": null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := dockerConfig{}
			if err := json.Unmarshal([]byte(tt.config), &cfg); err != nil {
				t.Fatalf("decoding test config: %v", err)
			}

			err, doNotUpdate := cfg.RemoveRuntime("amd")

			assert.NoError(t, err)
			assert.Equal(t, tt.wantDoNotUpdate, doNotUpdate, "doNotUpdate")

			want := dockerConfig{}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatalf("decoding expected config: %v", err)
			}
			assert.Equal(t, want, cfg)
		})
	}
}
