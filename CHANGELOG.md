# Changelog

All notable changes to **breathe** are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] - 2026-04-18

### Fixed

- Homebrew install no longer trips macOS Gatekeeper. The cask now
  strips `com.apple.quarantine` from the installed binary in a
  `postflight` hook, so `breathe` runs out of the box. Existing
  v0.1.0 installs can fix themselves locally with
  `xattr -dr com.apple.quarantine "$(brew --prefix breathe)"`, or
  just `brew upgrade breathe` once v0.1.1 is published.

## [0.1.0] - 2026-04-18

First public release.

### Added

- Configurable Pomodoro cycle: `--work`, `--short`, `--long`, `--rounds`.
- Spring-driven progress bar that tracks the real wall-clock elapsed time
  with sub-second precision (no per-second staircase).
- Phase-color morphing on transitions, driven by an RGB spring.
- Pulsing active session dot and a celebratory ripple flash on each
  just-completed dot, with the flanking separators briefly lighting up.
- Ambient outer-ring ripple around the main frame: two nested borders
  fade between near-background and the phase color as a pulse sweeps
  outward and back on a 10-second breath cycle.
- Guided breathing coach during break phases: an instruction
  ("breathe in" / "breathe out") and a fixed 8-slot dot grid whose
  brightness blooms outward from the center on the inhale and contracts
  on the exhale, synced to the same breath cycle as the outer rings.
- Help overlay (`?`) with a live mini-timer (phase, countdown, visual,
  dots) plus a key-and-flag reference. All controls remain live while
  the overlay is open.
- Notifications on phase boundaries: terminal bell by default,
  overridable with `--sound NAME` (macOS built-in) or `--bell-cmd CMD`
  (arbitrary shell command). Silenceable with `--no-bell`.
- Persistent session log under `$XDG_DATA_HOME/breathe/state.json`
  (atomic writes), used to count "today" and feed the stats subcommand.
- `breathe stats` subcommand that aggregates today and the rolling last
  seven days and renders the report through Glamour.
- Session-complete screen showing this-session, today, and last-7-days
  block + focus-time totals.
- Cross-platform release pipeline (GoReleaser + GitHub Actions) that
  builds darwin / linux / windows binaries on every `v*` tag and
  publishes a Homebrew cask via `Jaace/homebrew-tap`.

[Unreleased]: https://github.com/Jaace/breathe/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/Jaace/breathe/releases/tag/v0.1.1
[0.1.0]: https://github.com/Jaace/breathe/releases/tag/v0.1.0
