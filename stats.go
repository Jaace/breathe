package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
)

func runStats(args []string) {
	fs := flag.NewFlagSet("breathe stats", flag.ExitOnError)
	plain := fs.Bool("plain", false, "print plain text instead of rendered markdown")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	store, err := OpenStore()
	if err != nil {
		fmt.Fprintln(os.Stderr, "breathe: could not open store:", err)
		os.Exit(1)
	}

	md := buildStatsMarkdown(store.Sessions(), time.Now())

	if *plain {
		fmt.Println(md)
		return
	}

	out, err := glamour.Render(md, "auto")
	if err != nil {
		fmt.Println(md)
		return
	}
	fmt.Print(out)
}

// buildStatsMarkdown aggregates sessions into today + last-7-days summaries.
// `now` is a parameter for testability.
func buildStatsMarkdown(sessions []SessionRecord, now time.Time) string {
	y, m, d := now.Date()
	startToday := time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	startWeek := startToday.AddDate(0, 0, -6) // inclusive: 7 days incl today
	endWeek := startToday.Add(24 * time.Hour)

	todayWork := 0
	todayFocusSec := 0
	weekWork := 0
	weekFocusSec := 0
	byDay := map[string]int{}

	for _, r := range sessions {
		if r.Phase != PhaseWork.String() {
			continue
		}
		if !r.TS.Before(startToday) && r.TS.Before(endWeek) {
			todayWork++
			todayFocusSec += r.DurationSec
		}
		if !r.TS.Before(startWeek) && r.TS.Before(endWeek) {
			weekWork++
			weekFocusSec += r.DurationSec
			key := r.TS.Format("2006-01-02")
			byDay[key]++
		}
	}

	var b strings.Builder
	fmt.Fprintln(&b, "# breathe — stats")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Today")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- **%d** focus block(s)\n", todayWork)
	fmt.Fprintf(&b, "- **%s** focused time\n", formatMinutes(todayFocusSec))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Last 7 days")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- **%d** focus block(s)\n", weekWork)
	fmt.Fprintf(&b, "- **%s** focused time\n", formatMinutes(weekFocusSec))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| day | focus blocks |")
	fmt.Fprintln(&b, "|-----|--------------|")
	for i := 6; i >= 0; i-- {
		day := startToday.AddDate(0, 0, -i)
		key := day.Format("2006-01-02")
		fmt.Fprintf(&b, "| %s | %d |\n", day.Format("Mon Jan 2"), byDay[key])
	}
	return b.String()
}

func formatMinutes(sec int) string {
	if sec <= 0 {
		return "0m"
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
