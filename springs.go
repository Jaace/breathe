package main

import "github.com/charmbracelet/harmonica"

// FrameRate is the spring tick rate. Frame messages target this interval.
const FrameRate = 60

// Four named spring configurations. Each one is tuned for a specific visual
// job; changing a value here is the single knob to retune that behavior.

// ProgressSpring: critically damped. The filled ratio chases the real
// elapsed ratio smoothly, with no overshoot. Momentum means pause/resume
// eases instead of snapping.
func ProgressSpring() harmonica.Spring {
	return harmonica.NewSpring(harmonica.FPS(FrameRate), 6.0, 1.0)
}

// TransitionSpring: underdamped, snappy. Used on phase boundaries to let the
// progress bar spring back with a gentle overshoot before settling at zero.
func TransitionSpring() harmonica.Spring {
	return harmonica.NewSpring(harmonica.FPS(FrameRate), 4.0, 0.5)
}

// PulseSpring: slow, lightly damped. Drives the active session dot's size /
// brightness, gently oscillating around the target.
func PulseSpring() harmonica.Spring {
	return harmonica.NewSpring(harmonica.FPS(FrameRate), 2.0, 0.3)
}

// DigitSpring: fast, bouncy. Drives the y-offset for the "pomodoros today"
// digit flip so the new digit springs in from below.
func DigitSpring() harmonica.Spring {
	return harmonica.NewSpring(harmonica.FPS(FrameRate), 8.0, 0.4)
}

// ColorSpring: critically damped, used independently on R, G, B channels so
// the palette morphs smoothly across phase transitions.
func ColorSpring() harmonica.Spring {
	return harmonica.NewSpring(harmonica.FPS(FrameRate), 4.0, 1.0)
}

// ScalarTracker wraps a single-valued spring (position + velocity + target).
// Useful for any 1D animated value.
type ScalarTracker struct {
	Spring harmonica.Spring
	Pos    float64
	Vel    float64
	Target float64
}

func NewScalarTracker(s harmonica.Spring, initial float64) ScalarTracker {
	return ScalarTracker{Spring: s, Pos: initial, Target: initial}
}

func (t *ScalarTracker) SetTarget(target float64) { t.Target = target }

func (t *ScalarTracker) Tick() {
	t.Pos, t.Vel = t.Spring.Update(t.Pos, t.Vel, t.Target)
}

// RGBTracker animates an RGB triple by running three independent scalar
// springs in lockstep. Each channel keeps its own position and velocity.
type RGBTracker struct {
	R, G, B ScalarTracker
}

func NewRGBTracker(initial RGB) RGBTracker {
	s := ColorSpring()
	return RGBTracker{
		R: NewScalarTracker(s, initial.R),
		G: NewScalarTracker(s, initial.G),
		B: NewScalarTracker(s, initial.B),
	}
}

func (t *RGBTracker) SetTarget(c RGB) {
	t.R.SetTarget(c.R)
	t.G.SetTarget(c.G)
	t.B.SetTarget(c.B)
}

func (t *RGBTracker) Tick() {
	t.R.Tick()
	t.G.Tick()
	t.B.Tick()
}

func (t *RGBTracker) Current() RGB {
	return RGB{R: t.R.Pos, G: t.G.Pos, B: t.B.Pos}
}
