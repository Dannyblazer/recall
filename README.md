# rec — persistent, searchable shell history

## Build

```
go mod tidy   # fetches modernc.org/sqlite (pure Go, no cgo needed)
go build -o rec ./cmd/rec
sudo mv rec /usr/local/bin/
```

## Setup

```
rec init
```

Then add the relevant hook to your shell config:

```
# ~/.bashrc
source /path/to/hooks/rec.bash

# ~/.zshrc
source /path/to/hooks/rec.zsh

# ~/.config/fish/config.fish
source /path/to/hooks/rec.fish
```

Open a new shell (or `source` the config) and every command you run will be
logged in the background to `~/.local/share/rec/history.db`.

## Search

```
rec search psql                       # full-text search
rec search "docker run" --repo yemert # scoped to a git repo
rec search --failed --since 720h      # failed commands, last 30 days
```

## Status / not yet built

- `rec search` output is plain text for now — could add `--json` for scripting.
- No sync/backup yet. The `synced_at` column exists specifically so a future
  `rec sync` command can push unsynced rows to S3/a small server without a
  schema change.
- No dedup/pruning policy yet — worth deciding whether near-duplicate commands
  should collapse in search results or stay as distinct rows (they currently
  stay distinct, unlike bash's `HISTCONTROL=ignoreboth`).
