// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
)

func TestTrimExt(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"config.yaml", "config"},
		{"config.yml", "config"},
		{"config.json", "config"},
		{"config.tar.gz", "config.tar"},
		{"config", "config"},
		{"path/to/config.yaml", "path/to/config"},
		{".hidden", ".hidden"},
		{"", ""},
	}
	for _, tc := range cases {
		ext := trimExtString(tc.input)
		if ext != tc.want {
			t.Errorf("trimExtString(%q) = %q, want %q", tc.input, ext, tc.want)
		}
	}
}
