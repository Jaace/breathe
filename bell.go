package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Bell encapsulates what happens on a phase change notification. Zero value
// is disabled. The Ring method is non-blocking: external commands are
// launched async with Start so the TUI frame loop never stalls.
type Bell struct {
	Enabled bool   // master switch; false = complete silence
	Cmd     string // arbitrary shell command, overrides Sound if set
	Sound   string // friendly name (glass, ping, tink, ...) -> afplay on macOS
}

// macSounds maps friendly names to the built-in macOS alert sounds under
// /System/Library/Sounds. Keys are case-insensitive (normalized in Ring).
var macSounds = map[string]string{
	"basso":     "Basso",
	"blow":      "Blow",
	"bottle":    "Bottle",
	"frog":      "Frog",
	"funk":      "Funk",
	"glass":     "Glass",
	"hero":      "Hero",
	"morse":     "Morse",
	"ping":      "Ping",
	"pop":       "Pop",
	"purr":      "Purr",
	"sosumi":    "Sosumi",
	"submarine": "Submarine",
	"tink":      "Tink",
}

// MacSoundNames returns the sorted list of valid --sound values. Used for
// help text and validation.
func MacSoundNames() []string {
	out := make([]string, 0, len(macSounds))
	for k := range macSounds {
		out = append(out, k)
	}
	// Simple insertion sort; list is tiny and stdlib sort.Strings would do
	// but avoiding the import keeps this file dependency-free.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

// Ring fires the notification. Dispatch order:
//
//  1. disabled -> no-op
//  2. --bell-cmd -> shell out to it, async
//  3. --sound on darwin -> afplay the corresponding .aiff, async
//  4. fallback -> emit ASCII BEL (\a) to stdout
//
// A bad custom command or missing afplay degrades to BEL rather than
// surfacing an error; losing a beep mid-pomodoro isn't worth a popup.
func (b Bell) Ring() {
	if !b.Enabled {
		return
	}
	if b.Cmd != "" {
		if runCmd := shellRun(b.Cmd); runCmd != nil {
			if err := runCmd.Start(); err == nil {
				go runCmd.Wait()
				return
			}
		}
		fallbackBeep()
		return
	}
	if b.Sound != "" && runtime.GOOS == "darwin" {
		if file, ok := macSounds[strings.ToLower(b.Sound)]; ok {
			c := exec.Command("afplay", fmt.Sprintf("/System/Library/Sounds/%s.aiff", file))
			if err := c.Start(); err == nil {
				go c.Wait()
				return
			}
		}
	}
	fallbackBeep()
}

func fallbackBeep() { fmt.Print("\a") }

// shellRun wraps `cmd` in the platform's shell so pipes, quoting, and env
// expansion all work. Returns nil if the OS isn't supported.
func shellRun(cmd string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/C", cmd)
	default:
		return exec.Command("sh", "-c", cmd)
	}
}
