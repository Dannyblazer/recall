# recall — persistent, searchable shell history

## Install

### Debian / Ubuntu (apt)

```bash
curl -1sLf 'https://dl.cloudsmith.io/public/debian-o02u/recall/setup.deb.sh' | sudo -E bash
sudo apt-get install recall
```

This adds the signed `recall` apt repository and installs the latest release. Packages are built, signed, and published automatically via CI on every tagged release.

Note: Always remember "apt update" to get latest releases

### Build from source

```
go mod tidy   # fetches modernc.org/sqlite (pure Go, no cgo needed)
go build -o recall ./cmd/recall
```

## Setup

```
recall init
```

The necesary hook gets added to your shell config to catch/store all your commands.


Open a new shell (or `source` the config) and every command you run will be
logged in the background to `~/.local/share/recall/history.db`.

## Search

```
recall search psql                       # full-text search
recall search "docker run" --repo yemert # scoped to a git repo
recall search --failed --since 720h      # failed commands, last 30 days
```

## Status / not yet built

- `recall search` output is plain text for now — could add `--json` for scripting.
- No sync/backup yet. The `synced_at` column exists specifically so a future
  `recall sync` command can push unsynced rows to S3/a small server without a
  schema change.
- No dedup/pruning policy yet — worth deciding whether near-duplicate commands
  should collapse in search results or stay as distinct rows (they currently
  stay distinct, unlike bash's `HISTCONTROL=ignoreboth`).