package main

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
)

// RGB is a floating-point RGB triple in the 0..255 range. Floating point so
// springs can interpolate smoothly without rounding jitter.
type RGB struct {
	R, G, B float64
}

func (c RGB) Clamp() RGB {
	return RGB{
		R: clamp(c.R, 0, 255),
		G: clamp(c.G, 0, 255),
		B: clamp(c.B, 0, 255),
	}
}

func (c RGB) Hex() string {
	cc := c.Clamp()
	return fmt.Sprintf("#%02x%02x%02x",
		int(math.Round(cc.R)),
		int(math.Round(cc.G)),
		int(math.Round(cc.B)),
	)
}

// Mix blends c toward other by t in [0,1].
func (c RGB) Mix(other RGB, t float64) RGB {
	t = clamp(t, 0, 1)
	return RGB{
		R: c.R + (other.R-c.R)*t,
		G: c.G + (other.G-c.G)*t,
		B: c.B + (other.B-c.B)*t,
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Phase palettes. Warm red for focus; teal for the short break; cool blue for
// the long break.
var (
	ColorWork  = RGB{R: 224, G: 108, B: 117} // #E06C75
	ColorShort = RGB{R: 102, G: 194, B: 165} // #66C2A5
	ColorLong  = RGB{R: 101, G: 151, B: 213} // #6597D5

	ColorDim  = RGB{R: 90, G: 90, B: 100}
	ColorText = RGB{R: 220, G: 220, B: 230}

	// ColorBreathOff is the near-background color an outer breathing
	// ring fades to when the ripple pulse is far from it. Tuned for
	// dark terminals so an unlit ring is effectively invisible.
	ColorBreathOff = RGB{R: 24, G: 24, B: 30}
)

func PalettFor(kind PhaseKind) RGB {
	switch kind {
	case PhaseWork:
		return ColorWork
	case PhaseShortBreak:
		return ColorShort
	case PhaseLongBreak:
		return ColorLong
	}
	return ColorText
}

// Shared styles. The phase color is injected per-frame via Foreground, so
// these stay stateless.
var (
	styleLabel = lipgloss.NewStyle().
			Bold(true).
			Align(lipgloss.Center).
			Padding(0, 2)

	styleCountdown = lipgloss.NewStyle().
			Bold(true).
			Align(lipgloss.Center)

	styleBarTrack = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3a3a44"))

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDim.Hex()))

	styleFrame = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 3).
			Align(lipgloss.Center)

	styleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7a7a85")).
			Align(lipgloss.Center).
			PaddingTop(1)
)
