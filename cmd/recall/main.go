package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/dannyblazer/recall/internal/store"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	dbPath, err := defaultDBPath()
	if err != nil {
		fatal("resolve db path: %v", err)
	}

	switch os.Args[1] {
	case "init":
		cmdInit(dbPath)
	case "log":
		cmdLog(dbPath, os.Args[2:])
	case "search":
		cmdSearch(dbPath, os.Args[2:])
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// defaultDBPath resolves to ~/.local/share/recall/history.db, creating the
// directory if needed. XDG_DATA_HOME is respected if set.
func defaultDBPath() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		base = filepath.Join(u.HomeDir, ".local", "share")
	}
	dir := filepath.Join(base, "recall")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.db"), nil
}

func cmdInit(dbPath string) {
	s, err := store.Open(dbPath)
	if err != nil {
		fatal("init: %v", err)
	}
	defer s.Close()
	fmt.Printf("initialized history db at %s\n", dbPath)
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "recall: "+format+"\n", args...)
	os.Exit(1)
}

func printUsage() {
	fmt.Print(`recall - searchable, persistent shell command history

Usage:
  recall init                     Create the history database
  recall log [flags]              record one command (called by shell hooks)
  recall search <query> [flags]   Search command history

Run 'recall log -h' or 'recall search -h' for flag details.
`)
}
