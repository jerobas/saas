package main

import "testing"

func TestDevelopmentModeBypassesRemoteLicenseForLocalDevelopment(t *testing.T) {
	t.Setenv("SAAS_DEV_MODE", "true")
	active, err := getUserStatus()
	if err != nil {
		t.Fatal(err)
	}
	if !active {
		t.Fatal("development mode should unlock the local desktop application")
	}
}
