# breathe — VHS demo tape
#
# Re-record the demo gif:   vhs docs/demo.tape
# (writes to docs/demo.gif by default thanks to the Output below)
#
# Style notes:
#   - 90x28 leaves enough vertical room for the bordered frame +
#     outer ripple rings + footer without scroll/clip.
#   - JetBrains Mono is widely available and renders nerd-font-style
#     glyphs (●, ┌, etc.) cleanly. Fall back to your monospace of choice.
#   - The shot list packs work → break → second work → long break into
#     ~30s by aggressively shortening durations. Real pomodoros are
#     25/5/15; here we use 6/4/4 just for the recording.

Output docs/demo.gif

Set Theme        "Catppuccin Mocha"
Set FontFamily   "JetBrains Mono"
Set FontSize     14
Set Width        900
Set Height       560
Set Padding      24
Set TypingSpeed  60ms
Set PlaybackSpeed 1.0
Set Shell        zsh

# A clean prompt so the gif feels like a fresh terminal, not your dotfiles.
Hide
Type "export PS1='%F{8}❯%f '"
Enter
Type "clear"
Enter
Show

# Settle for one beat so viewers see the empty prompt before action.
Sleep 600ms

# Phase durations are chosen so the recording lands at least one full
# breath cycle (10s — see BreathCycleSeconds in timer.go) inside each
# phase. With --rounds 2 the sequence is: work → short → work → long →
# session-complete. Sleeps run ~1s longer than each phase so the cross-
# fade transition (color morph + dot ripple) is captured before the
# next phase's frame is also shown.
Type "breathe --work 10s --short 14s --long 12s --rounds 2"
Enter

# Phase 1 — WORK (10s) -----------------------------------------------
# Full breath cycle visible: outer rings sweep outward and back, spring
# progress bar fills smoothly, active session dot pulses, color stays
# in the work palette throughout.
Sleep 11s

# Phase 2 — SHORT BREAK (14s) ----------------------------------------
# The headline shot. Coach swaps in for the progress bar; instruction
# alternates "breathe in" / "breathe out"; dot grid blooms outward on
# the inhale and contracts on the exhale. 14s covers ~1.4 breath
# cycles so the viewer sees both directions clearly.
Sleep 15s

# Phase 3 — WORK (10s) -----------------------------------------------
# Color morph back to the work palette. Watch the entire frame
# (border, bar, dots) lerp over the brief overshoot before settling.
Sleep 11s

# Phase 4 — LONG BREAK (12s) -----------------------------------------
# Long-break palette + coach again. Slightly shorter than the short
# break since by now the viewer knows what the coach does — this beat
# is mostly about showing the third color phase.
Sleep 13s

# Session-complete card (3s) -----------------------------------------
# Lets the stats card render and settle (this-session / today /
# last-7-days totals).
Sleep 3s

# Quit cleanly so the gif ends on a stable frame, not a half-rendered
# spring.
Type "q"

Sleep 1.2s
