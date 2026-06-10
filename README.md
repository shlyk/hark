# hark

[![CI](https://github.com/shlyk/hark/actions/workflows/ci.yml/badge.svg)](https://github.com/shlyk/hark/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A macOS CLI that lets AI agents get your attention: native notification
banners, optional text-to-speech, and a local log of everything sent.

> *Hark! — listen, pay attention.*

## Why

AI coding agents work for minutes at a time. Instead of watching the
terminal, let the agent ping you when the build finishes, tests fail, or it
needs your input:

```
hark send "Tests passed, ready for review" -t "claude" --sound Glass
```

## Install

```sh
go install github.com/shlyk/hark@latest
```

Or build from source:

```sh
git clone https://github.com/shlyk/hark && cd hark
go build -o hark . && mv hark /usr/local/bin/
```

Requires macOS (uses the built-in `osascript` and `say`).

## Usage

```sh
hark send "Build finished"                       # banner with default title
hark send "Tests failed" -t CI --sound Glass     # title + sound
hark send "Need your input" --say                # banner + spoken aloud
hark send "Build done" --smart                   # speak only if headphones are connected
hark say "Deploy is done" --voice Samantha       # speech only
hark history -n 20                               # what pinged you lately
hark history --json                              # machine-readable
hark doctor                                      # verify delivery works
hark skill                                       # install the agent skill (Claude Code + Codex)
```

Run `hark doctor` once after installing — it checks dependencies, sends a
test banner, and reminds you how to grant notification permission if no
banner appears.

## For AI agents

Call `hark send "<message>"` whenever the user should be interrupted: a long
task finished, input is needed, or an error requires attention.

- Non-zero exit code means delivery failed.
- Every notification is recorded in `~/.local/state/hark/history.jsonl`
  (`$XDG_STATE_HOME` is honored); `hark history --json` reads it back.
- Message text is passed safely — quotes, backslashes, and AppleScript in
  the message cannot break or inject anything.

hark ships an [agent skill](skill/SKILL.md) that teaches agents when and how
to notify you. Install it once:

```sh
hark skill                    # ~/.claude/skills + ~/.codex/skills (all your projects)
hark skill --project          # ./.claude/skills + ./.codex/skills (commit with a repo)
hark skill --agent claude     # only one agent
```

Existing files are overwritten — re-run `hark skill` after upgrading.

## Notes

- Delivery uses `osascript`; banners appear under Script Editor's identity
  and are not clickable (accepted v1 trade-off — a native UserNotifications
  helper may replace it later without changing the CLI).
- `--smart` checks the default audio output via `system_profiler` (~1–2 s):
  Bluetooth audio or a wired headphone jack counts as headphones. If
  detection fails it falls back to a silent banner.
- Sounds: any name from `/System/Library/Sounds`, e.g. Glass, Ping, Sosumi.

## License

[MIT](LICENSE)
