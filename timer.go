package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickFrameMsg time.Time

func tickFrame() tea.Cmd {
	return tea.Tick(time.Second/time.Duration(FrameRate), func(t time.Time) tea.Msg { return tickFrameMsg(t) })
}

type model struct {
	// config
	cfg    SessionConfig
	phases []Phase
	bell   Bell

	// real clock. elapsed is a monotonic accumulator updated every frame
	// from the real wall-clock delta (dt) since the last tick. Sub-second
	// precision means the progress spring's target changes continuously,
	// not in 1-second steps, which removes the "staircase" feel on short
	// phases.
	phaseIdx int
	elapsed  time.Duration
	lastTick time.Time // zero before the first frame tick
	paused   bool
	finished bool

	// persistent state
	store        *Store
	todayCount   int // real count of completed work phases today
	displayCount int // discrete digit currently shown (lags behind todayCount)

	// springs
	progress ScalarTracker
	pulse    ScalarTracker
	digit    ScalarTracker
	color    RGBTracker

	// pulse control
	pulseHi         bool
	pulseNextToggle time.Time

	// transient transition state
	transitioning bool

	// UI
	width    int
	height   int
	showHelp bool
}

func runBubbleTea(cfg SessionConfig, bell Bell) error {
	phases := BuildSequence(cfg)
	if len(phases) == 0 {
		return fmt.Errorf("empty phase sequence")
	}

	store, err := OpenStore()
	if err != nil {
		return err
	}
	today := store.CountToday(PhaseWork)

	initialColor := PalettFor(phases[0].Kind)

	m := model{
		cfg:          cfg,
		phases:       phases,
		bell:         bell,
		store:        store,
		todayCount:   today,
		displayCount: today,
		progress:     NewScalarTracker(ProgressSpring(), 0),
		pulse:        NewScalarTracker(PulseSpring(), 1.0),
		digit:        NewScalarTracker(DigitSpring(), 1.0),
		color:        NewRGBTracker(initialColor),
	}
	m.pulseNextToggle = time.Now().Add(900 * time.Millisecond)
	m.pulse.SetTarget(1.15)

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return tickFrame()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// All controls stay live whether the help overlay is open or not;
		// the timer keeps ticking either way. Only `q` is context sensitive:
		// it closes the overlay when open, and quits otherwise.
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case " ", "space":
			m.paused = !m.paused
			return m, nil
		case "s":
			return m.advancePhase(true), nil
		case "r":
			m.elapsed = 0
			m.progress.SetTarget(0)
			return m, nil
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "esc":
			if m.showHelp {
				m.showHelp = false
			}
			return m, nil
		case "q":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			return m, tea.Quit
		}
		return m, nil

	case tickFrameMsg:
		now := time.Time(msg)

		// Accumulate wall-clock elapsed for the current phase using the
		// real dt since the last frame. First tick after Init has a zero
		// lastTick, so dt is 0 and elapsed doesn't jump.
		var dt time.Duration
		if !m.lastTick.IsZero() {
			dt = now.Sub(m.lastTick)
		}
		m.lastTick = now

		if !m.finished {
			advanced := false
			if !m.paused {
				m.elapsed += dt
				if m.elapsed >= m.phases[m.phaseIdx].Duration {
					m = m.advancePhase(false)
					advanced = true
				}
			}

			// Progress target tracks the real fractional elapsed every
			// frame. During a phase-boundary transition the target is
			// temporarily driven below zero (see advancePhase) to produce
			// the overshoot; we release that mode once Pos has settled.
			// Skip the target/release logic on the same frame we just
			// advanced so the overshoot target (-0.08) stays in effect.
			if !advanced {
				if !m.transitioning {
					phaseLen := m.phases[m.phaseIdx].Duration
					ratio := float64(m.elapsed) / float64(phaseLen)
					if ratio > 1 {
						ratio = 1
					}
					m.progress.SetTarget(ratio)
				} else if m.progress.Pos < 0.02 {
					m.transitioning = false
					m.progress.SetTarget(0)
				}
			}
			m.progress.Tick()

			// Pulse: toggle the target between high and low to keep the
			// underdamped spring oscillating indefinitely on the active dot.
			if now.After(m.pulseNextToggle) {
				m.pulseHi = !m.pulseHi
				if m.pulseHi {
					m.pulse.SetTarget(1.15)
				} else {
					m.pulse.SetTarget(0.85)
				}
				m.pulseNextToggle = now.Add(900 * time.Millisecond)
			}
			m.pulse.Tick()

			// Digit: if the real count is ahead of the displayed one, pop
			// the digit down (Pos -> 0) so the View knows to render the new
			// digit sliding up.
			if m.displayCount < m.todayCount {
				if m.digit.Pos > 0.98 && m.digit.Target > 0.5 {
					m.digit.SetTarget(0.0)
				}
				if m.digit.Pos < 0.05 && m.digit.Target < 0.5 {
					m.displayCount++
					m.digit.SetTarget(1.0)
				}
			}
			m.digit.Tick()

			// Color follows the current phase's palette.
			m.color.Tick()
		}
		return m, tickFrame()
	}

	return m, nil
}

// advancePhase moves to the next phase in the sequence. skipped=true means
// the user hit `s` (don't record, don't bell). Returns the updated model.
func (m model) advancePhase(skipped bool) model {
	prev := m.phases[m.phaseIdx]

	if !skipped {
		if prev.Kind == PhaseWork {
			if err := m.store.Record(prev.Kind, prev.Duration); err == nil {
				m.todayCount++
			}
		}
		m.bell.Ring()
	}

	m.phaseIdx++
	if m.phaseIdx >= len(m.phases) {
		m.finished = true
		m.phaseIdx = len(m.phases) - 1
		m.elapsed = m.phases[m.phaseIdx].Duration
		return m
	}

	m.elapsed = 0
	m.transitioning = true
	// Brief negative target produces the Harmonica overshoot that snaps
	// the bar back below zero before settling. The frame handler clears
	// transitioning once Pos returns near 0.
	m.progress.SetTarget(-0.08)
	m.color.SetTarget(PalettFor(m.phases[m.phaseIdx].Kind))
	return m
}
