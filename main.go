package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// version is injected by GoReleaser at build time via -ldflags.
var version = "dev"

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "stats":
			runStats(os.Args[2:])
			return
		case "version", "--version", "-v":
			fmt.Println("breathe", version)
			return
		case "help", "--help", "-h":
			printUsage(os.Stdout)
			return
		}
	}
	runTimer(os.Args[1:])
}

func printUsage(w *os.File) {
	fmt.Fprintf(w, `breathe — a Pomodoro timer with physics

Usage:
  breathe [flags]         start a session
  breathe stats           show today + last-7-days totals
  breathe --version       print version

Flags:
  --work       duration of a work block       (default 25m)
  --short      duration of a short break      (default 5m)
  --long       duration of the long break     (default 15m)
  --rounds     work blocks before long break  (default 4)
  --no-bell    suppress the terminal bell

Keys:
  space  pause/resume    s  skip    r  reset    q  quit    ?  help
`)
}

func runTimer(args []string) {
	fs := flag.NewFlagSet("breathe", flag.ExitOnError)
	fs.Usage = func() { printUsage(os.Stderr) }
	work := fs.Duration("work", 25*time.Minute, "duration of a work block")
	short := fs.Duration("short", 5*time.Minute, "duration of a short break")
	long := fs.Duration("long", 15*time.Minute, "duration of the long break")
	rounds := fs.Int("rounds", 4, "number of work blocks before the long break")
	noBell := fs.Bool("no-bell", false, "suppress the terminal bell on phase change")
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if *showVersion {
		fmt.Println("breathe", version)
		return
	}
	if *rounds < 1 {
		fmt.Fprintln(os.Stderr, "breathe: --rounds must be >= 1")
		os.Exit(2)
	}
	if *work <= 0 || *short <= 0 || *long <= 0 {
		fmt.Fprintln(os.Stderr, "breathe: durations must be > 0")
		os.Exit(2)
	}

	cfg := SessionConfig{
		Work:   *work,
		Short:  *short,
		Long:   *long,
		Rounds: *rounds,
	}
	if err := runBubbleTea(cfg, !*noBell); err != nil {
		fmt.Fprintln(os.Stderr, "breathe:", err)
		os.Exit(1)
	}
}
