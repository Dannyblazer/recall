package store

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GitContext returns (repoName, branch) for the given directory, or ("", "")
// if it's not inside a git repo. Shells out to `git` rather than parsing
// .git internals directly, since that stays correct across git versions.
func GitContext(cwd string) (repo string, branch string) {
	top, err := runGit(cwd, "rev-parse", "--show-toplevel")
	if err != nil || top == "" {
		return "", ""
	}
	repo = filepath.Base(top)

	branch, err = runGit(cwd, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		branch = ""
	}
	return repo, branch
}

func runGit(cwd string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
