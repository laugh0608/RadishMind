package main

import "testing"

func TestReadOptionalBooleanEnvironment(t *testing.T) {
	const name = "RADISHMIND_APPLICATION_RESULT_ARTIFACT_LIBRARY_DEV_FIXTURE_TEST"
	t.Setenv(name, "")
	if enabled, err := readOptionalBooleanEnvironment(name); err != nil || enabled {
		t.Fatalf("empty optional boolean drifted: enabled=%v err=%v", enabled, err)
	}
	t.Setenv(name, " true ")
	if enabled, err := readOptionalBooleanEnvironment(name); err != nil || !enabled {
		t.Fatalf("true optional boolean drifted: enabled=%v err=%v", enabled, err)
	}
	t.Setenv(name, "sometimes")
	if enabled, err := readOptionalBooleanEnvironment(name); err == nil || enabled {
		t.Fatalf("invalid optional boolean was accepted: enabled=%v err=%v", enabled, err)
	}
}
