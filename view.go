package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const columnWidth = 52

func (m model) View() string {
	if m.finished {
		return m.renderFinished()
	}

	phase := m.phases[m.phaseIdx]
	phaseColor := m.color.Current().Hex()

	label := styleLabel.
		Foreground(lipgloss.Color(phaseColor)).
		Render(phase.Kind.Label())

	remaining := phase.Duration - m.elapsed
	if remaining < 0 {
		remaining = 0
	}
	countdown := styleCountdown.
		Foreground(lipgloss.Color(phaseColor)).
		Render(formatCountdown(remaining))

	bar := m.renderBar(phaseColor)
	dots := m.renderDots()
	footer := m.renderFooter(phaseColor)

	pausedNote := ""
	if m.paused {
		pausedNote = styleDim.Render("— paused —")
	}

	column := lipgloss.JoinVertical(lipgloss.Center,
		label,
		"",
		countdown,
		pausedNote,
		"",
		bar,
		"",
		dots,
		"",
		footer,
	)

	framed := styleFrame.
		BorderForeground(lipgloss.Color(phaseColor)).
		Width(columnWidth).
		Render(column)

	help := m.renderHelp()

	block := lipgloss.JoinVertical(lipgloss.Center, framed, help)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
	}
	return block
}

func (m model) renderFinished() string {
	title := styleLabel.
		Foreground(lipgloss.Color(ColorLong.Hex())).
		Render("SESSION COMPLETE")
	msg := styleDim.Render(fmt.Sprintf("%d focused block(s) today. Press q to quit.", m.todayCount))

	block := lipgloss.JoinVertical(lipgloss.Center, title, "", msg)
	framed := styleFrame.
		BorderForeground(lipgloss.Color(ColorLong.Hex())).
		Width(columnWidth).
		Render(block)
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, framed)
	}
	return framed
}

func formatCountdown(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Round(time.Second).Seconds())
	m := total / 60
	s := total % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// renderBar draws a spring-filled progress bar. Width is fixed; characters
// flip from "track" to "fill" based on the spring position. We also render a
// fractional leading character so sub-cell motion is visible (that's where
// the Harmonica smoothness shows up even on a narrow bar).
func (m model) renderBar(phaseColor string) string {
	width := columnWidth - 8
	if width < 10 {
		width = 10
	}
	pos := m.progress.Pos
	if pos < 0 {
		pos = 0
	}
	if pos > 1 {
		pos = 1
	}
	floatFill := pos * float64(width)
	fullCells := int(floatFill)
	frac := floatFill - float64(fullCells)

	// Sub-cell shading characters (0..1 fill)
	partials := []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉'}
	partialIdx := int(frac * float64(len(partials)-1))
	if partialIdx < 0 {
		partialIdx = 0
	}
	if partialIdx >= len(partials) {
		partialIdx = len(partials) - 1
	}

	fillStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(phaseColor))
	trackStyle := styleBarTrack

	var b strings.Builder
	for i := 0; i < fullCells && i < width; i++ {
		b.WriteString(fillStyle.Render("█"))
	}
	if fullCells < width {
		b.WriteString(fillStyle.Render(string(partials[partialIdx])))
		for i := fullCells + 1; i < width; i++ {
			b.WriteString(trackStyle.Render("─"))
		}
	}
	return b.String()
}

// renderDots draws one glyph per phase. Completed phases are dim, the active
// dot pulses via the spring (by swapping glyph at pulse thresholds), and
// upcoming phases are tracked as hollow.
func (m model) renderDots() string {
	var b strings.Builder
	for i, ph := range m.phases {
		glyph := "○"
		styled := styleDim
		switch {
		case i < m.phaseIdx:
			glyph = "●"
			styled = lipgloss.NewStyle().Foreground(lipgloss.Color(PalettFor(ph.Kind).Hex()))
		case i == m.phaseIdx:
			activeColor := m.color.Current().Hex()
			if m.pulse.Pos >= 1.0 {
				glyph = "●"
			} else {
				glyph = "◉"
			}
			styled = lipgloss.NewStyle().
				Foreground(lipgloss.Color(activeColor)).
				Bold(true)
		default:
			glyph = "○"
			styled = styleDim
		}
		b.WriteString(styled.Render(glyph))
		if i < len(m.phases)-1 {
			b.WriteString(styleDim.Render(" "))
		}
	}
	return b.String()
}

// renderFooter shows the "today" counter. The active digit is rendered with a
// vertical offset driven by the digit spring so the new digit slides up.
func (m model) renderFooter(phaseColor string) string {
	label := styleDim.Render("today")
	digit := m.renderDigit(fmt.Sprintf("%d", m.displayCount), phaseColor)
	return lipgloss.JoinHorizontal(lipgloss.Center, label, "  ", digit)
}

// renderDigit renders a number with a 2-line vertical slot so the spring's
// y-offset is visible. Pos ~ 1 = digit centered; Pos ~ 0 = digit just below.
func (m model) renderDigit(s, phaseColor string) string {
	slot := lipgloss.NewStyle().
		Foreground(lipgloss.Color(phaseColor)).
		Bold(true).
		Width(len(s) + 2).
		Align(lipgloss.Center)
	top := ""
	bot := ""
	if m.digit.Pos > 0.5 {
		top = slot.Render(s)
		bot = ""
	} else {
		top = ""
		bot = slot.Render(s)
	}
	return lipgloss.JoinVertical(lipgloss.Center, top, bot)
}

func (m model) renderHelp() string {
	if m.showHelp {
		return styleHelp.Render("space pause  ·  s skip  ·  r reset  ·  q quit  ·  ? close help")
	}
	return styleHelp.Render("space pause  ·  s skip  ·  r reset  ·  q quit  ·  ? help")
}
