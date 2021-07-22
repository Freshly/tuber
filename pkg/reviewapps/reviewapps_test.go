package reviewapps

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testCases = []struct {
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
		name:     "too long only alphanumeric - 76 characters",
		input:    "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
		expected: "abcdefghijklmnopqrstuvwxyzabcd",
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
		name:     "leading hyphens",
		input:    "--foo",
		expected: "foo",
	},
	{
		name:     "trailing hyphens",
		input:    "foo--",
		expected: "foo",
	},
	{
		name:     "only hyphens",
		input:    "----",
		expected: fmt.Sprintf("review-app-%d", time.Now().Unix()),
	},
}

func TestMakeDNS1123Compatible(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := makeDNS1123Compatible(tc.input)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

// Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^(BenchmarkMakeDNS1123Compatible)$ github.com/freshly/tuber/pkg/reviewapps

// goos: darwin
// goarch: amd64
// pkg: github.com/freshly/tuber/pkg/reviewapps
// cpu: Intel(R) Core(TM) i7-8850H CPU @ 2.60GHz
// BenchmarkMakeDNS1123Compatible/only_numbers-12         	 4625367	       258.2 ns/op	      16 B/op	       2 allocs/op
// BenchmarkMakeDNS1123Compatible/only_lowercase_letters_and_hyphens-12         	 3029067	       397.7 ns/op	      32 B/op	       2 allocs/op
// BenchmarkMakeDNS1123Compatible/capital_letters-12                            	 3096699	       389.1 ns/op	      16 B/op	       3 allocs/op
// BenchmarkMakeDNS1123Compatible/underscores_to_hyphens-12                     	 1376887	       851.2 ns/op	      48 B/op	       3 allocs/op
// BenchmarkMakeDNS1123Compatible/capitals_and_underscores-12                   	 1000000	      1023 ns/op	      64 B/op	       4 allocs/op
// BenchmarkMakeDNS1123Compatible/too_long_only_alphanumeric_-_76_characters-12 	 1525465	       801.8 ns/op	     112 B/op	       2 allocs/op
// BenchmarkMakeDNS1123Compatible/symbols-12                                    	  597916	      1888 ns/op	      72 B/op	       5 allocs/op
// BenchmarkMakeDNS1123Compatible/symbols_without_modifying_valid_hyphens-12    	  645900	      1869 ns/op	      72 B/op	       5 allocs/op
// BenchmarkMakeDNS1123Compatible/leading_hyphens-12                            	 1732497	       703.4 ns/op	      40 B/op	       3 allocs/op
// BenchmarkMakeDNS1123Compatible/trailing_hyphens-12                           	 1244368	       987.9 ns/op	      40 B/op	       3 allocs/op
// BenchmarkMakeDNS1123Compatible/only_hyphens-12                               	 1671430	       686.3 ns/op	      64 B/op	       4 allocs/op
// PASS
// ok  	github.com/freshly/tuber/pkg/reviewapps	18.334s
func BenchmarkMakeDNS1123Compatible(b *testing.B) {
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				makeDNS1123Compatible(tc.input)
			}
		})
	}
}
