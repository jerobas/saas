package main

import (
	"path/filepath"
	"testing"
)

func TestDataDirectoryUsesConfiguredPath(t *testing.T) {
	expected := filepath.Join(t.TempDir(), "desktop-data")
	t.Setenv("SAAS_DATA_DIR", expected)

	if actual := dataDirectory(); actual != expected {
		t.Fatalf("dataDirectory() = %q, want %q", actual, expected)
	}
}
