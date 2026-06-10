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
hark send "<short message>" -t "<source>"       # banner; -t names the sender, e.g. claude, codex, ci
hark send "<message>" -t claude --sound Glass   # + sound (any name from /System/Library/Sounds)
hark send "<message>" -t claude --say           # + spoken aloud
```

- Keep messages under ~100 characters; lead with the outcome ("Tests passed",
  "Build failed: 3 errors", "Need your decision on X").
- Default to a plain banner; add `--sound` for failures or blocking questions.
  Use `--say` only when the user said they are stepping away.
- Any message text is safe to pass — hark escapes quotes and special characters itself.

## Verification

- Exit code 0 = delivered; non-zero = delivery failed (stderr explains).
  Attempt at most twice in total, then report the failure in your response instead.
