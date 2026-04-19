# breathe

A Pomodoro timer with physics. Every transition — progress bars filling, countdown digits morphing, phase changes, dot pulses — runs through [Harmonica](https://github.com/charmbracelet/harmonica) springs, so nothing snaps or ticks abruptly. The whole thing breathes.

[**Live site & demo →**](https://jaace.github.io/breathe/)

![A short recording of breathe running: spring-driven progress bar, breathing rings rippling outward from the frame, color-morphing phase transitions, the breathing coach during a short break, and the session-complete stats card.](docs/demo.gif)

## Install

```bash
brew install Jaace/tap/breathe
```

Or grab a prebuilt binary from the [releases page](https://github.com/Jaace/breathe/releases).

## Usage

```bash
breathe                 # default 25/5 cycle, 4 rounds, long break at the end
breathe --work 50m --short 10m --long 30m --rounds 3
breathe --no-bell       # silent transitions
breathe --sound glass   # macOS built-in sound (glass, ping, tink, hero, ...)
breathe --bell-cmd "afplay ~/sounds/ding.wav"   # any shell command
breathe stats           # today + last-7-days totals
breathe --version
```

### Keys

| key          | action             |
| ------------ | ------------------ |
| `space`      | pause / resume     |
| `s`          | skip current phase |
| `r`          | reset current phase|
| `q` / `ctrl+c` | quit             |
| `?`          | toggle help        |

## Flags

| flag         | default | meaning                                              |
| ------------ | ------- | ---------------------------------------------------- |
| `--work`     | `25m`   | duration of a work block                             |
| `--short`    | `5m`    | duration of a short break                            |
| `--long`     | `15m`   | duration of the long break                           |
| `--rounds`   | `4`     | work blocks before the long break                    |
| `--no-bell`        | off     | suppress the notification sound entirely             |
| `--sound`          | —       | macOS built-in sound name (see below)                |
| `--bell-cmd`       | —       | arbitrary shell command to run on phase change       |
| `--no-update-check`| off     | skip the once-per-day check for a newer release      |

Durations accept anything [`time.ParseDuration`](https://pkg.go.dev/time#ParseDuration) takes (`90s`, `1h30m`, etc.).

### Notification sound

By default, `breathe` emits a terminal bell (`\a`) on phase change — what that actually sounds like depends on your terminal's bell settings. To pick a specific sound:

- `--sound NAME` plays a built-in macOS alert via `afplay`. Valid names: `basso`, `blow`, `bottle`, `frog`, `funk`, `glass`, `hero`, `morse`, `ping`, `pop`, `purr`, `sosumi`, `submarine`, `tink`. Silently falls back to the terminal bell on non-macOS systems.
- `--bell-cmd CMD` runs any shell command you want. Works on Linux (`paplay file.ogg`), macOS (`afplay /path/to/file.wav`), Windows (`powershell ...`), or whatever else. Overrides `--sound` if both are set.
- `--no-bell` silences everything.

## Data

Session history lives at `$XDG_DATA_HOME/breathe/state.json` (falls back to `~/.local/share/breathe/state.json`). Append-only, one entry per completed phase.

## Update checks

Once a day, `breathe` quietly asks the GitHub API whether there's a newer release. If there is, the footer shows a small `vX.Y.Z available · brew upgrade breathe` notice. Results are cached at `$XDG_CACHE_HOME/breathe/update-check.json` (falls back to `~/.cache/breathe/`) so the network is hit at most once every 24 hours. Any failure (offline, rate-limited, etc.) is silent. Pass `--no-update-check` to disable the lookup entirely.

## Build from source

```bash
git clone https://github.com/Jaace/breathe
cd breathe
go build -o breathe .
```

## License

MIT
