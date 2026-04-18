package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type SessionRecord struct {
	TS          time.Time `json:"ts"`
	Phase       string    `json:"phase"`
	DurationSec int       `json:"duration_sec"`
}

type StoreData struct {
	Sessions []SessionRecord `json:"sessions"`
}

type Store struct {
	path string
	data StoreData
}

// DataDir returns the directory used for persistent data, respecting
// XDG_DATA_HOME. Always returns a non-empty string; the caller should
// MkdirAll before writing.
func DataDir() (string, error) {
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "breathe"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "breathe"), nil
}

// OpenStore loads (or initializes) the JSON store.
func OpenStore() (*Store, error) {
	dir, err := DataDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	p := filepath.Join(dir, "state.json")

	s := &Store{path: p}
	bs, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return s, nil
		}
		return nil, err
	}
	if len(bs) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(bs, &s.data); err != nil {
		// corrupt file: back it up and start fresh rather than fail-hard.
		_ = os.Rename(p, p+".corrupt."+time.Now().Format("20060102-150405"))
		s.data = StoreData{}
	}
	return s, nil
}

// Record appends a phase completion and persists atomically.
func (s *Store) Record(kind PhaseKind, duration time.Duration) error {
	s.data.Sessions = append(s.data.Sessions, SessionRecord{
		TS:          time.Now(),
		Phase:       kind.String(),
		DurationSec: int(duration.Round(time.Second).Seconds()),
	})
	return s.flush()
}

func (s *Store) flush() error {
	bs, err := json.MarshalIndent(&s.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, bs, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// CountToday returns the number of completed sessions of `kind` in the
// local-day window containing now.
func (s *Store) CountToday(kind PhaseKind) int {
	now := time.Now()
	y, m, d := now.Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	n := 0
	for _, r := range s.data.Sessions {
		if r.Phase != kind.String() {
			continue
		}
		if (r.TS.Equal(start) || r.TS.After(start)) && r.TS.Before(end) {
			n++
		}
	}
	return n
}

// Sessions returns a read-only snapshot of all recorded sessions.
func (s *Store) Sessions() []SessionRecord {
	return s.data.Sessions
}

// String helper (for debugging).
func (r SessionRecord) String() string {
	return fmt.Sprintf("%s %s %ds", r.TS.Format(time.RFC3339), r.Phase, r.DurationSec)
}
