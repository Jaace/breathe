package main

import "time"

type PhaseKind int

const (
	PhaseWork PhaseKind = iota
	PhaseShortBreak
	PhaseLongBreak
)

func (p PhaseKind) String() string {
	switch p {
	case PhaseWork:
		return "work"
	case PhaseShortBreak:
		return "short"
	case PhaseLongBreak:
		return "long"
	}
	return "unknown"
}

func (p PhaseKind) Label() string {
	switch p {
	case PhaseWork:
		return "FOCUS"
	case PhaseShortBreak:
		return "SHORT BREAK"
	case PhaseLongBreak:
		return "LONG BREAK"
	}
	return ""
}

type Phase struct {
	Kind     PhaseKind
	Duration time.Duration
}

type SessionConfig struct {
	Work   time.Duration
	Short  time.Duration
	Long   time.Duration
	Rounds int
}

func DefaultConfig() SessionConfig {
	return SessionConfig{
		Work:   25 * time.Minute,
		Short:  5 * time.Minute,
		Long:   15 * time.Minute,
		Rounds: 4,
	}
}

// BuildSequence produces a classic pomodoro cycle:
//
//	work, short, work, short, ..., work, long
//
// with `Rounds` work blocks, `Rounds-1` short breaks between them, and one
// long break at the end. For Rounds=4 this is the canonical
// 25/5/25/5/25/5/25/15.
func BuildSequence(cfg SessionConfig) []Phase {
	if cfg.Rounds < 1 {
		cfg.Rounds = 1
	}
	out := make([]Phase, 0, cfg.Rounds*2)
	for i := 0; i < cfg.Rounds; i++ {
		out = append(out, Phase{Kind: PhaseWork, Duration: cfg.Work})
		if i < cfg.Rounds-1 {
			out = append(out, Phase{Kind: PhaseShortBreak, Duration: cfg.Short})
		}
	}
	out = append(out, Phase{Kind: PhaseLongBreak, Duration: cfg.Long})
	return out
}
