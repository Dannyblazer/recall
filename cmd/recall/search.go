package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/dannyblazer/recall/internal/store"
)

func cmdSearch(dbPath string, args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	repo := fs.String("repo", "", "filter by git repo name")
	cwd := fs.String("cwd", "", "filter by working directory prefix")
	failed := fs.Bool("failed", false, "only show commands that exited non-zero")
	since := fs.String("since", "", "only show commands after this time, e.g. \"2026-01-01\" or \"720h\" (30 days)")
	limit := fs.Int("limit", 50, "max results")
	fs.Parse(args)

	query := ""
	if fs.NArg() > 0 {
		query = fs.Arg(0)
	}

	var sinceTime time.Time
	if *since != "" {
		t, err := parseSince(*since)
		if err != nil {
			fatal("search: invalid --since value: %v", err)
		}
		sinceTime = t
	}

	s, err := store.Open(dbPath)
	if err != nil {
		fatal("search: %v", err)
	}
	defer s.Close()

	results, err := s.Search(store.SearchOptions{
		Query:      query,
		Repo:       *repo,
		Cwd:        *cwd,
		FailedOnly: *failed,
		Since:      sinceTime,
		Limit:      *limit,
	})
	if err != nil {
		fatal("search: %v", err)
	}

	for _, c := range results {
		marker := " "
		if c.ExitCode != 0 {
			marker = "x"
		}
		fmt.Printf("[%s] %s  %-20s  %s\n",
			marker,
			c.StartedAt.Format("2006-01-02 15:04"),
			shortRepo(c.GitRepo, c.Cwd),
			c.Command,
		)
	}
}

func shortRepo(repo, cwd string) string {
	if repo != "" {
		return repo
	}
	return cwd
}

// parseSince accepts either a Go duration (e.g. "720h") applied as "ago",
// or an absolute date in YYYY-MM-DD form.
func parseSince(s string) (time.Time, error) {
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(-d), nil
	}
	return time.Parse("2006-01-02", s)
}
