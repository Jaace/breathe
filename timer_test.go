package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel(t *testing.T) model {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)
	store, err := OpenStore()
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	cfg := SessionConfig{
		Work:   2 * time.Second,
		Short:  1 * time.Second,
		Long:   3 * time.Second,
		Rounds: 2,
	}
	phases := BuildSequence(cfg)
	return model{
		cfg:      cfg,
		phases:   phases,
		bell:     Bell{Enabled: false},
		store:    store,
		progress: NewScalarTracker(ProgressSpring(), 0),
		pulse:    NewScalarTracker(PulseSpring(), 1.0),
		digit:    NewScalarTracker(DigitSpring(), 1.0),
		color:    NewRGBTracker(PalettFor(phases[0].Kind)),
	}
}

func TestBuildSequenceDefault(t *testing.T) {
	cfg := DefaultConfig()
	seq := BuildSequence(cfg)
	// 4 rounds: W S W S W S W L -> 8 phases
	if len(seq) != 8 {
		t.Fatalf("expected 8 phases, got %d", len(seq))
	}
	expected := []PhaseKind{
		PhaseWork, PhaseShortBreak,
		PhaseWork, PhaseShortBreak,
		PhaseWork, PhaseShortBreak,
		PhaseWork, PhaseLongBreak,
	}
	for i, p := range seq {
		if p.Kind != expected[i] {
			t.Errorf("phase %d: got %v want %v", i, p.Kind, expected[i])
		}
	}
}

func TestBuildSequenceSingleRound(t *testing.T) {
	cfg := SessionConfig{Work: time.Minute, Short: time.Minute, Long: time.Minute, Rounds: 1}
	seq := BuildSequence(cfg)
	if len(seq) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(seq))
	}
	if seq[0].Kind != PhaseWork || seq[1].Kind != PhaseLongBreak {
		t.Errorf("unexpected shape: %+v", seq)
	}
}

func TestPauseResumeDoesNotAdvanceClock(t *testing.T) {
	m := newTestModel(t)
	m.paused = true
	for i := 0; i < 5; i++ {
		next, _ := m.Update(tickSecondMsg(time.Now()))
		m = next.(model)
	}
	if m.elapsed != 0 {
		t.Errorf("paused clock advanced: %v", m.elapsed)
	}
}

func TestClockAdvancesAndAdvancesPhase(t *testing.T) {
	m := newTestModel(t)
	// Work duration is 2s in test config; 2 ticks should complete the phase.
	for i := 0; i < 2; i++ {
		next, _ := m.Update(tickSecondMsg(time.Now()))
		m = next.(model)
	}
	if m.phaseIdx != 1 {
		t.Errorf("expected phaseIdx 1 after work completes, got %d", m.phaseIdx)
	}
	if m.elapsed != 0 {
		t.Errorf("expected elapsed reset to 0 on phase advance, got %v", m.elapsed)
	}
	if m.todayCount != 1 {
		t.Errorf("expected todayCount 1 after work phase, got %d", m.todayCount)
	}
	if !m.transitioning {
		t.Errorf("expected transitioning=true immediately after phase advance")
	}
}

func TestSkipDoesNotRecordOrCountToday(t *testing.T) {
	m := newTestModel(t)
	m = m.advancePhase(true)
	if m.todayCount != 0 {
		t.Errorf("skip should not increment todayCount, got %d", m.todayCount)
	}
}

func TestResetKey(t *testing.T) {
	m := newTestModel(t)
	// advance elapsed a bit
	next, _ := m.Update(tickSecondMsg(time.Now()))
	m = next.(model)
	if m.elapsed == 0 {
		t.Fatal("elapsed should be non-zero after one tick")
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = next.(model)
	if m.elapsed != 0 {
		t.Errorf("reset should zero elapsed, got %v", m.elapsed)
	}
}

func TestStoreRecordPersists(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)
	s, err := OpenStore()
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	if err := s.Record(PhaseWork, 25*time.Minute); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "breathe", "state.json")); err != nil {
		t.Fatalf("state file not written: %v", err)
	}
	s2, err := OpenStore()
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if n := s2.CountToday(PhaseWork); n != 1 {
		t.Errorf("CountToday=%d want 1", n)
	}
}

func TestFullCycleCompletes(t *testing.T) {
	m := newTestModel(t)
	// Sum of durations = 2 (W) + 1 (S) + 2 (W) + 3 (L) = 8 seconds.
	for i := 0; i < 10; i++ {
		next, _ := m.Update(tickSecondMsg(time.Now()))
		m = next.(model)
	}
	if !m.finished {
		t.Errorf("expected session finished, still at phaseIdx=%d", m.phaseIdx)
	}
	if m.todayCount != 2 {
		t.Errorf("expected 2 work phases recorded, got %d", m.todayCount)
	}
}
