package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeriveTelemetryURL(t *testing.T) {
	tests := []struct {
		name     string
		apiURL   string
		expected string
	}{
		{
			name:     "empty string",
			apiURL:   "",
			expected: "",
		},
		{
			name:     "invalid URL",
			apiURL:   "://bad",
			expected: "",
		},
		{
			name:     "URL with no host",
			apiURL:   "not-a-url",
			expected: "",
		},
		{
			name:     "production master",
			apiURL:   "https://api.cast.ai",
			expected: "telemetry.prod-master.cast.ai",
		},
		{
			name:     "production master with port",
			apiURL:   "https://api.cast.ai:443",
			expected: "telemetry.prod-master.cast.ai",
		},
		{
			name:     "production master with path",
			apiURL:   "https://api.cast.ai/v1/foo",
			expected: "telemetry.prod-master.cast.ai",
		},
		{
			name:     "dev-master env",
			apiURL:   "https://api.dev-master.cast.ai",
			expected: "telemetry.dev-master.cast.ai",
		},
		{
			name:     "EU production",
			apiURL:   "https://api.eu.cast.ai",
			expected: "telemetry.prod-eu.cast.ai",
		},
		{
			name:     "local dev URL",
			apiURL:   "https://api--myenv.local.cast.ai",
			expected: "api-grpc--myenv.local.cast.ai",
		},
		{
			name:     "local dev URL with wrong prefix",
			apiURL:   "https://notapi.local.cast.ai",
			expected: "",
		},
		{
			name:     "local dev URL with extra parts",
			apiURL:   "https://api--x.y.local.cast.ai",
			expected: "",
		},
		{
			name:     "non-cast.ai domain",
			apiURL:   "https://api.example.com",
			expected: "",
		},
		{
			name:     "too many host parts",
			apiURL:   "https://api.a.b.cast.ai",
			expected: "",
		},
		{
			name:     "host not starting with api",
			apiURL:   "https://web.cast.ai",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			r.Equal(tt.expected, deriveTelemetryURL(tt.apiURL))
		})
	}
}
