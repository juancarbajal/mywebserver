package main

import (
	"os"
	"testing"
)

func TestOpenBrowserEnvironmentVariable(t *testing.T) {
	// Set BROWSER to a dummy non-existent command to test that BROWSER env var is honored
	orig := os.Getenv("BROWSER")
	defer os.Setenv("BROWSER", orig)

	os.Setenv("BROWSER", "echo")
	err := openBrowser("http://localhost:8000")
	if err != nil {
		t.Fatalf("expected openBrowser to succeed with BROWSER=echo, got: %v", err)
	}
}
