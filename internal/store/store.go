package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Store wraps the SQLite connection used for all command history reads/writes.
type Store struct {
	db *sql.DB
}

// Command represents a single captured shell command.
type Command struct {
	ID         int64
	Command    string
	Cwd        string
	GitRepo    string
	GitBranch  string
	ExitCode   int
	DurationMs int64
	Shell      string
	Hostname   string
	SessionID  string
	StartedAt  time.Time
}

// Open opens (or creates) the SQLite database at path and ensures the schema exists.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// Insert records one captured command.
func (s *Store) Insert(c Command) error {
	_, err := s.db.Exec(`
		INSERT INTO commands
			(command, cwd, git_repo, git_branch, exit_code, duration_ms, shell, hostname, session_id, started_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Command, c.Cwd, c.GitRepo, c.GitBranch, c.ExitCode, c.DurationMs,
		c.Shell, c.Hostname, c.SessionID, c.StartedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("insert command: %w", err)
	}
	return nil
}

// SearchOptions filters a history search.
type SearchOptions struct {
	Query      string // FTS5 match query; empty means no text filter
	Repo       string
	Cwd        string
	FailedOnly bool
	Since      time.Time // zero value means no lower bound
	Limit      int
}

// Search returns commands matching the given options, most recent first.
func (s *Store) Search(opts SearchOptions) ([]Command, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT c.id, c.command, c.cwd, c.git_repo, c.git_branch, c.exit_code,
		       c.duration_ms, c.shell, c.hostname, c.session_id, c.started_at
		FROM commands c`
	args := []any{}
	conditions := []string{}

	if opts.Query != "" {
		query += ` JOIN commands_fts f ON f.rowid = c.id`
		conditions = append(conditions, `commands_fts MATCH ?`)
		args = append(args, opts.Query)
	}
	if opts.Repo != "" {
		conditions = append(conditions, `c.git_repo = ?`)
		args = append(args, opts.Repo)
	}
	if opts.Cwd != "" {
		conditions = append(conditions, `c.cwd LIKE ? || '%'`)
		args = append(args, opts.Cwd)
	}
	if opts.FailedOnly {
		conditions = append(conditions, `c.exit_code != 0`)
	}
	if !opts.Since.IsZero() {
		conditions = append(conditions, `c.started_at >= ?`)
		args = append(args, opts.Since.Unix())
	}

	for i, cond := range conditions {
		if i == 0 {
			query += " WHERE " + cond
		} else {
			query += " AND " + cond
		}
	}
	query += " ORDER BY c.started_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()

	var results []Command
	for rows.Next() {
		var c Command
		var startedAt int64
		if err := rows.Scan(
			&c.ID, &c.Command, &c.Cwd, &c.GitRepo, &c.GitBranch, &c.ExitCode,
			&c.DurationMs, &c.Shell, &c.Hostname, &c.SessionID, &startedAt,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		c.StartedAt = time.Unix(startedAt, 0)
		results = append(results, c)
	}
	return results, rows.Err()
}
