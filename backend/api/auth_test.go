package api

import "testing"

func TestSafeNextPath(t *testing.T) {
	cases := map[string]bool{
		"/":                     true,
		"/s/abc123":             true,
		"/tracks/4?tab=x":       true,
		"":                      false,
		"s/abc":                 false,
		"//evil.example":        false,
		"/\\evil.example":       false,
		"https://evil.example":  false,
		"/ok\r\nSet-Cookie: x=": false,
	}
	for in, want := range cases {
		if got := safeNextPath(in); got != want {
			t.Errorf("safeNextPath(%q) = %v, want %v", in, got, want)
		}
	}
}
