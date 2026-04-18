package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const columnWidth = 52

func (m model) View() string {
	if m.showHelp {
		return m.renderHelpOverlay()
	}
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
	accent := ColorLong.Hex()

	title := styleLabel.
		Foreground(lipgloss.Color(accent)).
		Render("SESSION COMPLETE")
	kudos := styleDim.Render("nicely done")

	// Session stats come straight from the config (all N blocks completed
	// by definition once we hit this screen). Today and week come from the
	// store so they reflect persisted history across runs.
	sessionBlocks := m.cfg.Rounds
	sessionFocus := time.Duration(sessionBlocks) * m.cfg.Work

	now := time.Now()
	y, mo, d := now.Date()
	startToday := time.Date(y, mo, d, 0, 0, 0, 0, now.Location())
	startWeek := startToday.AddDate(0, 0, -6)
	endWeek := startToday.Add(24 * time.Hour)

	todayBlocks, todaySec := 0, 0
	weekBlocks, weekSec := 0, 0
	for _, r := range m.store.Sessions() {
		if r.Phase != PhaseWork.String() {
			continue
		}
		if !r.TS.Before(startToday) && r.TS.Before(endWeek) {
			todayBlocks++
			todaySec += r.DurationSec
		}
		if !r.TS.Before(startWeek) && r.TS.Before(endWeek) {
			weekBlocks++
			weekSec += r.DurationSec
		}
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorDim.Hex())).
		Width(14).
		Align(lipgloss.Right)
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(accent)).
		Bold(true).
		PaddingLeft(2)

	row := func(label, value string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render(label), valueStyle.Render(value))
	}

	stats := lipgloss.JoinVertical(lipgloss.Left,
		row("this session", fmt.Sprintf("%d blocks · %s", sessionBlocks, formatFocus(sessionFocus))),
		row("today", fmt.Sprintf("%d blocks · %s", todayBlocks, formatFocus(time.Duration(todaySec)*time.Second))),
		row("last 7 days", fmt.Sprintf("%d blocks · %s", weekBlocks, formatFocus(time.Duration(weekSec)*time.Second))),
	)

	hint := styleHelp.Render("press q to quit")

	block := lipgloss.JoinVertical(lipgloss.Center,
		title,
		kudos,
		"",
		stats,
		"",
		hint,
	)
	framed := styleFrame.
		BorderForeground(lipgloss.Color(accent)).
		Width(columnWidth).
		Render(block)
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, framed)
	}
	return framed
}

// formatFocus renders a Duration as a compact "1h 25m" / "45m" / "30s"
// string suitable for the completion screen.
func formatFocus(d time.Duration) string {
	if d <= 0 {
		return "0m"
	}
	total := int(d.Round(time.Second).Seconds())
	h := total / 3600
	m := (total % 3600) / 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	case m > 0:
		return fmt.Sprintf("%dm", m)
	default:
		return fmt.Sprintf("%ds", total)
	}
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

// renderFooter shows the "today" counter as a vertically stacked block:
// the digit sits on top, the "today" label below, both centered. When the
// digit spring is mid-flip (Pos < 0.5) the digit row is briefly blanked out
// so the new value appears to spring in from below.
func (m model) renderFooter(phaseColor string) string {
	digitStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(phaseColor)).
		Bold(true)
	digitRow := digitStyle.Render(fmt.Sprintf("%d", m.displayCount))
	if m.digit.Pos < 0.5 {
		digitRow = " "
	}
	label := styleDim.Render("today")
	return lipgloss.JoinVertical(lipgloss.Center, digitRow, label)
}

func (m model) renderHelp() string {
	return styleHelp.Render("space pause  ·  s skip  ·  r reset  ·  q quit  ·  ? help")
}

// renderMiniTimer is a compact header used at the top of the help overlay
// so the user doesn't lose sight of the active session: phase + countdown,
// a narrow progress bar, and session dots.
func (m model) renderMiniTimer(phaseColor string) string {
	if m.finished {
		return styleDim.Render("session complete")
	}

	phase := m.phases[m.phaseIdx]
	remaining := phase.Duration - m.elapsed
	if remaining < 0 {
		remaining = 0
	}

	phaseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(phaseColor)).
		Bold(true)

	parts := []string{
		phaseStyle.Render(phase.Kind.Label()),
		styleDim.Render("·"),
		phaseStyle.Render(formatCountdown(remaining)),
	}
	if m.paused {
		parts = append(parts, styleDim.Render("·"), styleDim.Render("paused"))
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center, interleaveSpaces(parts)...)
	bar := m.renderBar(phaseColor)
	dots := m.renderDots()

	return lipgloss.JoinVertical(lipgloss.Center, header, "", bar, "", dots)
}

// interleaveSpaces puts a single-space separator between every element.
func interleaveSpaces(parts []string) []string {
	if len(parts) <= 1 {
		return parts
	}
	out := make([]string, 0, len(parts)*2-1)
	for i, p := range parts {
		if i > 0 {
			out = append(out, " ")
		}
		out = append(out, p)
	}
	return out
}

// renderHelpOverlay replaces the main view with a full keybinding / flag
// reference. A minimal timer header stays visible at the top so the user
// still knows where they are in the session. Dismissed with `?`, `q`, or
// `esc`.
func (m model) renderHelpOverlay() string {
	phaseColor := m.color.Current().Hex()

	miniTimer := m.renderMiniTimer(phaseColor)

	title := styleLabel.
		Foreground(lipgloss.Color(phaseColor)).
		Render("HELP")

	// Single consistent layout for both tables: right-aligned key column,
	// fixed-width desc column, so KEYS and FLAGS share the exact same
	// internal grid and feel tidy when stacked.
	const (
		keyColW  = 11
		descColW = 28
	)
	keyCol := lipgloss.NewStyle().
		Foreground(lipgloss.Color(phaseColor)).
		Bold(true).
		Width(keyColW).
		Align(lipgloss.Right)
	descCol := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText.Hex())).
		Width(descColW).
		PaddingLeft(2)

	row := func(key, desc string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top, keyCol.Render(key), descCol.Render(desc))
	}

	keys := lipgloss.JoinVertical(lipgloss.Left,
		row("space", "pause / resume"),
		row("s", "skip current phase"),
		row("r", "reset current phase"),
		row("?", "toggle this help"),
		row("q, ctrl+c", "quit"),
	)

	sectionTitle := func(s string) string {
		return styleDim.Bold(true).Render(s)
	}

	flagsBlock := lipgloss.JoinVertical(lipgloss.Left,
		row("--work", "work block (25m)"),
		row("--short", "short break (5m)"),
		row("--long", "long break (15m)"),
		row("--rounds", "blocks per cycle (4)"),
		row("--no-bell", "silence notifications"),
		row("--sound", "macOS built-in sound"),
		row("--bell-cmd", "custom shell command"),
	)

	hint := styleHelp.Render("press ? or q to close")

	column := lipgloss.JoinVertical(lipgloss.Center,
		miniTimer,
		"",
		title,
		"",
		sectionTitle("KEYS"),
		"",
		keys,
		"",
		sectionTitle("FLAGS"),
		"",
		flagsBlock,
		"",
		hint,
	)

	framed := styleFrame.
		BorderForeground(lipgloss.Color(phaseColor)).
		Width(columnWidth).
		Render(column)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, framed)
	}
	return framed
}
