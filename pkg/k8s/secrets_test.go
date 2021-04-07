package k8s

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessData(t *testing.T) {
	testCases := []struct {
		name               string
		input              []byte
		envKeys, envValues []string
		expected           map[string]string
	}{
		{
			name:     "no data",
			expected: nil,
		},
		{
			name: "no env vars in config",
			input: []byte(`
FOO_DEBUG: "true"
FOO_ENVIRONMENT: "staging"
		`),
			expected: map[string]string{"FOO_DEBUG": encode("true"), "FOO_ENVIRONMENT": encode("staging")},
		},
		{
			name: "env vars present",
			input: []byte(`
FOOKEY: ${TEST_SECRET_ENV_KEY}
FOO_COUNT: ${TEST-COUNT-123}
`),
			envKeys:   []string{"TEST_SECRET_ENV_KEY", "TEST-COUNT-123"},
			envValues: []string{"env-key", "456"},
			expected:  map[string]string{"FOOKEY": encode("env-key"), "FOO_COUNT": encode("456")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setUnsetEnv(tc.envKeys, tc.envValues)

			actual := processData(tc.input)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func setUnsetEnv(keys, values []string) func() {
	for i, k := range keys {
		os.Setenv(k, values[i])
	}

	return func() {
		for _, k := range keys {
			os.Unsetenv(k)
		}
	}
}

func encode(in string) string {
	return base64.StdEncoding.EncodeToString([]byte(in))
}
