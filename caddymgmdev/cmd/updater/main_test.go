package main

import "testing"

func TestReleaseVersionPattern(t *testing.T) {
	valid := []string{"v0.10.3", "0.10.3", "v1.2"}
	invalid := []string{"", "latest", "v1.2.3;rm", "v1.2.3-beta", "../1.2.3"}
	for _, value := range valid {
		if !releaseVersionPattern.MatchString(value) {
			t.Errorf("expected %q to be accepted", value)
		}
	}
	for _, value := range invalid {
		if releaseVersionPattern.MatchString(value) {
			t.Errorf("expected %q to be rejected", value)
		}
	}
}
