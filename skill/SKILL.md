---
name: hark
description: Use when the user should be notified or asked something while away from the terminal — a long-running task finished, input or a decision is needed, a build/test/deploy completed or failed, or an error needs attention. Sends native macOS notifications and interactive question dialogs via the hark CLI.
---

# Notifying the user with hark

`hark` sends macOS notification banners and can ask questions via dialogs.
Use it to get the user's attention; do not assume they are watching the
terminal. macOS only — if `hark` is not on PATH or you are not on macOS,
skip notifying; do not attempt to install it.

## When to notify

- A task that ran longer than ~30 seconds just finished (build, tests, deploy, long analysis).
- You are blocked waiting for the user's input or a decision.
- Something failed and needs their attention.

Send ONE notification per state change — a build that stays broken is one
event, not one per check. In watcher loops always use a dedupe key:
`--once <stable-key>` skips repeats within 10 minutes (a skipped duplicate
still exits 0; do not retry without the key).

## How

```sh
hark send "<short message>" -t "<source>" --smart   # banner; spoken aloud only if the user wears headphones
hark send "<message>" -t claude --sound Glass       # + sound (any name from /System/Library/Sounds)
hark send "<message>" -t claude --say               # force speech even on speakers (only if the user asked to be spoken to)
hark ask "<question>" --timeout 300                 # Yes/No dialog; prints the answer to stdout
hark ask "<question>" --options "A,B,C"             # pick-one list (no timeout support); prints the choice
hark ask "<question>" --input --timeout 300         # free-text answer
```

- Default to `--smart` with `-t` naming you as the sender (e.g. claude, codex, ci).
  `--smart` speech is private (headphones only), so it needs no explicit consent;
  `--say` plays over speakers too. Flags combine freely.
- When blocked on a real decision, prefer `hark ask` over `hark send` — the
  user can answer from the dialog without returning to the terminal. Set
  `--timeout` so you never hang unattended. A non-zero exit from `ask` IS
  the user's answer ("not now") — never show a second dialog; fall back to
  asking in the terminal.
- Keep messages under ~100 characters; lead with the outcome ("Tests passed",
  "Build failed: 3 errors", "Need your decision on X").
- Any message text is safe to pass — hark escapes quotes and special characters itself.

## Verification

- Exit code 0 = delivered / answered; non-zero = failed (stderr explains).
- `send` delivery failure: attempt at most twice in total, then report the
  failure in your response instead. This retry rule is for `send` only —
  never re-prompt a dismissed or timed-out `ask`.
- `hark history -n 5` shows recent notifications if you suspect a duplicate.
