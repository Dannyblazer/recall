package main

import (
	"flag"
	"os"
	"time"

	"github.com/dannyblazer/recall/internal/store"
)

func cmdLog(dbPath string, args []string) {
	fs := flag.NewFlagSet("log", flag.ExitOnError)
	cmdText := fs.String("cmd", "", "the command text (required)")
	exitCode := fs.Int("exit", 0, "exit code of the command")
	durationMs := fs.Int64("duration", 0, "duration in milliseconds")
	shell := fs.String("shell", "", "shell name: bash | zsh | fish")
	sessionID := fs.String("session", "", "terminal session id")
	cwd := fs.String("cwd", "", "working directory (defaults to $PWD)")
	fs.Parse(args)

	if *cmdText == "" {
		fatal("log: --cmd is required")
	}

	dir := *cwd
	if dir == "" {
		dir, _ = os.Getwd()
	}

	hostname, _ := os.Hostname()
	repo, branch := store.GitContext(dir)

	s, err := store.Open(dbPath)
	if err != nil {
		fatal("log: %v", err)
	}
	defer s.Close()

	err = s.Insert(store.Command{
		Command:    *cmdText,
		Cwd:        dir,
		GitRepo:    repo,
		GitBranch:  branch,
		ExitCode:   *exitCode,
		DurationMs: *durationMs,
		Shell:      *shell,
		Hostname:   hostname,
		SessionID:  *sessionID,
		StartedAt:  time.Now(),
	})
	if err != nil {
		fatal("log: %v", err)
	}
}
