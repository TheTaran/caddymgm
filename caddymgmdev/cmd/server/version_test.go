package main

import "testing"

func TestNumericVersion(t *testing.T) {
	tests := []struct {
		value string
		want  [3]int
		ok    bool
	}{
		{"v2.11.4", [3]int{2, 11, 4}, true},
		{"0.10.2-dirty", [3]int{0, 10, 2}, true},
		{"v0.10", [3]int{0, 10, 0}, true},
		{"go1.27.0", [3]int{1, 27, 0}, true},
		{"development", [3]int{}, false},
		{"v1.2.3.4", [3]int{}, false},
	}
	for _, test := range tests {
		got, ok := numericVersion(test.value)
		if got != test.want || ok != test.ok {
			t.Errorf("numericVersion(%q) = %v, %v; want %v, %v", test.value, got, ok, test.want, test.ok)
		}
	}
}

func TestIsVersionNewer(t *testing.T) {
	tests := []struct {
		candidate string
		current   string
		want      bool
	}{
		{"v0.10.3", "v0.10.2", true},
		{"v0.11.0", "v0.10.9", true},
		{"v2.11.4", "v2.11.4", false},
		{"v0.10.2", "v0.10.2-dirty", false},
		{"v2.10.0", "v2.11.0", false},
		{"go1.27.0", "go1.26.7", true},
		{"invalid", "v1.0.0", false},
	}
	for _, test := range tests {
		if got := isVersionNewer(test.candidate, test.current); got != test.want {
			t.Errorf("isVersionNewer(%q, %q) = %v; want %v", test.candidate, test.current, got, test.want)
		}
	}
}
