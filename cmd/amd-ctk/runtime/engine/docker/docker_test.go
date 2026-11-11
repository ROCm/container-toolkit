package docker

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
