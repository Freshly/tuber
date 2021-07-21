package reviewapps

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeDNS1123Compatible(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "only numbers",
			input:    "12345",
			expected: "12345",
		},
		{
			name:     "only lowercase letters and hyphens",
			input:    "a-b-c-d-e-f",
			expected: "a-b-c-d-e-f",
		},
		{
			name:     "capital letters",
			input:    "FOO",
			expected: "foo",
		},
		{
			name:     "underscores to hyphens",
			input:    "1_2_3_4_5",
			expected: "1-2-3-4-5",
		},
		{
			name:     "capitals and underscores",
			input:    "A_B_C_D_E",
			expected: "a-b-c-d-e",
		},
		{
			name:     "too long only alphanumeric",
			input:    strings.Repeat("a", dns1123NameMaximumLength+1),
			expected: strings.Repeat("a", dns1123NameMaximumLength),
		},
		{
			name:     "symbols",
			input:    "foo&bar/foo@bar",
			expected: "foobarfoobar",
		},
		{
			name:     "symbols without modifying valid hyphens",
			input:    "foo&bar-foo@bar",
			expected: "foobar-foobar",
		},
		{
			name:     "trailing hyphen",
			input:    "foo-",
			expected: "foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := makeDNS1123Compatible(tc.input)

			assert.Equal(t, tc.expected, actual)
		})
	}
}
