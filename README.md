# hark

A macOS CLI that lets AI agents get your attention: native notification
banners, optional text-to-speech, and a local log of everything sent.

## Install

    go install .    # installs `hark` into $GOPATH/bin
    # or: go build -o hark . && mv hark /usr/local/bin/

## Usage

    hark send "Build finished"                       # banner with default title
    hark send "Tests failed" -t CI --sound Glass     # title + sound
    hark send "Need your input" --say                # banner + spoken aloud
    hark say "Deploy is done" --voice Samantha       # speech only
    hark history -n 20                               # what pinged you lately
    hark history --json                              # machine-readable
    hark doctor                                      # verify delivery works

## For AI agents

Call `hark send "<message>"` whenever the user should be interrupted:
long task finished, input needed, error requiring attention. Non-zero exit
means delivery failed. Notifications are recorded in
`~/.local/state/hark/history.jsonl`.

## Notes

- Delivery uses `osascript`; banners appear under Script Editor's identity
  and are not clickable (accepted v1 trade-off).
- Sounds: any name from /System/Library/Sounds, e.g. Glass, Ping, Sosumi.
