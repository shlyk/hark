---
name: hark
description: Use when the user should be notified while away from the terminal — a long-running task finished, input or a decision is needed, a build/test/deploy completed or failed, or an error needs attention. Sends a native macOS notification (optionally spoken aloud) via the hark CLI.
---

# Notifying the user with hark

`hark` sends macOS notification banners. Use it to get the user's attention;
do not assume they are watching the terminal. macOS only — if `hark` is not
on PATH or you are not on macOS, skip notifying; do not attempt to install it.

## When to notify

- A task that ran longer than ~30 seconds just finished (build, tests, deploy, long analysis).
- You are blocked waiting for the user's input or a decision.
- Something failed and needs their attention.

Send ONE notification per event. Never send notifications in a loop. If you
suspect you already notified about this event, check `hark history -n 5` first.

## How

```sh
hark send "<short message>" -t "<source>" --smart   # banner; spoken aloud only if the user wears headphones
hark send "<message>" -t claude --sound Glass       # + sound (any name from /System/Library/Sounds)
hark send "<message>" -t claude --say               # force speech even on speakers (only if the user asked to be spoken to)
```

- Default to `--smart` with `-t` naming you as the sender (e.g. claude, codex, ci).
  `--smart` speech is private (headphones only), so it needs no explicit consent;
  `--say` plays over speakers too. Flags combine freely.
- Keep messages under ~100 characters; lead with the outcome ("Tests passed",
  "Build failed: 3 errors", "Need your decision on X").
- Any message text is safe to pass — hark escapes quotes and special characters itself.

## Verification

- Exit code 0 = delivered; non-zero = delivery failed (stderr explains).
  Attempt at most twice in total, then report the failure in your response instead.
