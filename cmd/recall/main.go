package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/dannyblazer/recall/internal/store"
)

type shellType string

// hooksDir is where apt install the shell hooks scripts for "recall logs"
const hooksDir = "/usr/share/recall/hooks"

// hookMarker is written into each hook script so appendShellHook can detect
// whether it's already been added, instead of appending duplicates on every
// `recall init` run.
const hookMarker = "recall shell hook"

const (
	bashShell shellType = "bash"
	zshShell  shellType = "zsh"
	fishShell shellType = "fish"
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

	// add shell hooks to terminal source to route commands to recall logs
	// first get shell Type
	shell := detectShell()
	switch shell {
	case string(bashShell):
		// add to bash shell .bashrc
		appendShellHook("recall.bash", ".bashrc")
	case string(zshShell):
		// add to zsh shell
		appendShellHook("recall.zsh", ".zshrc")
	case string(fishShell):
		// add to fish shell
		appendShellHook("recall.fish", "config.fish")
	default:
		fatal("could not detect your shell from $SHELL — source the appropriate hook manually from %s", hooksDir)
	}

}

func appendShellHook(hookName, shellFile string) {
	// fetch home dir
	home, err := os.UserHomeDir()
	if err != nil {
		// return "", err
		fatal("unable to access homeDir: %v", err)
	}

	dir, err := resolveHooksDir()
	if err != nil {
		fatal("%v", err)
	}

	// Open source file first
	fullpath := filepath.Join(dir, hookName)
	src, err := os.Open(fullpath)
	if err != nil {
		fatal("failed to open source file: %v", err)
	}
	defer src.Close()

	dstPath := filepath.Join(home, shellFile)
	// Open destination file
	if shellFile == "config.fish" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			fatal("unable to get fish config dir: %v", err)
		}
		dstPath = filepath.Join(configDir, "fish", shellFile)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			fatal("unable to create fish config dir: %v", err)
		}
	}

	// Skip if the hook's already present, so re-running `recall init`
	// doesn't keep appending duplicate copies.
	if existing, err := os.ReadFile(dstPath); err == nil {
		if strings.Contains(string(existing), hookMarker) {
			fmt.Printf("recall hook already present in %s, skipping\n", dstPath)
			return
		}
	}

	dst, err := os.OpenFile(dstPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fatal("failed to open destination file: %v", err)
	}
	defer dst.Close()

	// stream content to shell file
	_, err = io.Copy(dst, src)
	if err != nil {
		fatal("unable to append content to shell config: %v", err)
	}
	fmt.Println("Setup Complete")
}

func detectShell() string {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return ""
	}
	return filepath.Base(shellPath)
}

func resolveHooksDir() (string, error) {
	if _, err := os.Stat(hooksDir); err == nil {
		return hooksDir, nil
	}

	exe, err := os.Executable()
	if err == nil {
		local := filepath.Join(filepath.Dir(exe), "hooks")
		if _, err := os.Stat(local); err == nil {
			return local, nil
		}
	}

	cwdHooks := filepath.Join(".", "hooks")
	if _, err := os.Stat(cwdHooks); err == nil {
		return cwdHooks, nil
	}

	return "", fmt.Errorf("could not find hook scripts in %s, next to the binary, or in ./hooks — if you built from source, run recall init from the repo root", hooksDir)
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
