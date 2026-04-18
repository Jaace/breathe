package main

import (
	"runtime"
	"testing"
)

func TestBellDisabledIsSilent(t *testing.T) {
	// No panic, no error; just verify the dispatch path is reachable.
	b := Bell{Enabled: false, Sound: "glass", Cmd: "false"}
	b.Ring()
}

func TestBellCustomCmdDispatches(t *testing.T) {
	// Use `true`, which exists on every unix-like system and exits 0.
	if runtime.GOOS == "windows" {
		t.Skip("custom cmd test uses sh")
	}
	b := Bell{Enabled: true, Cmd: "true"}
	b.Ring() // should not panic, returns before the goroutine's Wait.
}

func TestMacSoundNamesSorted(t *testing.T) {
	names := MacSoundNames()
	if len(names) != len(macSounds) {
		t.Fatalf("MacSoundNames returned %d entries, want %d", len(names), len(macSounds))
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("MacSoundNames not sorted: %q > %q at index %d", names[i-1], names[i], i)
		}
	}
}

func TestMacSoundsKnown(t *testing.T) {
	required := []string{"glass", "ping", "tink", "hero", "submarine"}
	for _, name := range required {
		if _, ok := macSounds[name]; !ok {
			t.Errorf("macSounds missing expected key %q", name)
		}
	}
}
