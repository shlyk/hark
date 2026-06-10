# hark

[![CI](https://github.com/shlyk/hark/actions/workflows/ci.yml/badge.svg)](https://github.com/shlyk/hark/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A macOS CLI that lets AI agents get your attention: native notification
banners, two-way question dialogs, text-to-speech, and a local log of
everything sent.

> *Hark! — listen, pay attention.*

## Why

AI coding agents work for minutes at a time. Instead of watching the
terminal, let the agent ping you when the build finishes, tests fail, or it
needs your input — and even answer its questions from a dialog:

```
$ hark ask "Tests pass. Deploy to staging?" --options "Deploy,Hold"
Deploy
```

## Install

```sh
go install github.com/shlyk/hark@latest
```

Or build from source: `go build -o hark . && mv hark /usr/local/bin/`.
Requires macOS (uses the built-in `osascript` and `say`).

## Usage

```sh
hark send "Build finished"                       # banner with default title
hark send "Tests failed" -t CI --sound Glass     # title + sound
hark send "Build done" --smart                   # speak only if headphones are connected
hark send "Need your input" --say                # banner + spoken aloud
hark send "Build OK" --once build-42             # dedupe: skip repeats within 10 min
hark ask "Deploy now?"                           # Yes/No dialog, answer on stdout
hark ask "Strategy?" --options A,B,C             # pick from a list
hark ask "Ticket id?" --input                    # free-text answer
hark say "Deploy is done" --voice Samantha       # speech only
hark history -n 20 [--json] [--follow]           # what pinged you lately
hark doctor                                      # verify the whole setup
hark skill                                       # install the agent skill (Claude Code + Codex)
hark hook claude                                 # auto-notify from Claude Code hooks
```

Run `hark doctor` once after installing — it checks dependencies and config,
sends a test banner, and reminds you how to grant notification permission if
no banner appears.

## Configuration (optional)

`~/.config/hark/config.json` — all fields optional; flags always win:

```json
{
  "title": "hark",
  "smart": true,
  "sound": "Glass"
}
```

## Claude Code integration

```sh
hark hook claude      # adds Notification + Stop hooks to ~/.claude/settings.json
hark hook claude --remove
```

After this, every Claude Code session notifies you when it needs your
attention (permission prompts, questions) and when it finishes. The previous
settings file is backed up to `settings.json.bak`.

## For AI agents

- `hark send "<msg>" --smart` to notify; `hark ask` to get decisions.
- Non-zero exit = delivery failed / question unanswered.
- Use `--once <key>` to avoid duplicate pings for the same event.
- Everything is recorded in `~/.local/state/hark/history.jsonl`
  (`$XDG_STATE_HOME` honored); `hark history --json` reads it back.
- Message text is passed safely — quotes, backslashes, and AppleScript in
  the message cannot break or inject anything.

hark ships an [agent skill](skill/SKILL.md) that teaches agents when and how
to notify you. `hark skill` installs it for Claude Code and Codex
(`--project` for the current repo, `--agent claude|codex` to filter).
Re-run after upgrading hark.

## Notes

- Delivery uses `osascript`; banners appear under Script Editor's identity
  and are not clickable (accepted trade-off — no .app bundle, no signing).
- `--smart` checks the default audio output via `system_profiler` (~1–2 s):
  Bluetooth audio or a wired headphone jack counts as headphones. Detection
  failure falls back to a silent banner.
- Everything stays on your machine — hark makes no network calls.
- Sounds: any name from `/System/Library/Sounds`, e.g. Glass, Ping, Sosumi.

## License

[MIT](LICENSE)
