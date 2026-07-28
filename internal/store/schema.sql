CREATE TABLE IF NOT EXISTS commands (
    id          INTEGER PRIMARY KEY,
    command     TEXT NOT NULL,
    cwd         TEXT,
    git_repo    TEXT,
    git_branch  TEXT,
    exit_code   INTEGER,
    duration_ms INTEGER,
    shell       TEXT,
    hostname    TEXT,
    session_id  TEXT,
    started_at  INTEGER NOT NULL,
    synced_at   INTEGER
);

CREATE INDEX IF NOT EXISTS idx_commands_started_at ON commands(started_at);
CREATE INDEX IF NOT EXISTS idx_commands_git_repo    ON commands(git_repo);
CREATE INDEX IF NOT EXISTS idx_commands_exit_code   ON commands(exit_code);

-- Full-text search index over the command text, kept in sync via triggers below.
CREATE VIRTUAL TABLE IF NOT EXISTS commands_fts USING fts5(
    command,
    content='commands',
    content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS commands_ai AFTER INSERT ON commands BEGIN
    INSERT INTO commands_fts(rowid, command) VALUES (new.id, new.command);
END;

CREATE TRIGGER IF NOT EXISTS commands_ad AFTER DELETE ON commands BEGIN
    INSERT INTO commands_fts(commands_fts, rowid, command) VALUES ('delete', old.id, old.command);
END;

CREATE TRIGGER IF NOT EXISTS commands_au AFTER UPDATE ON commands BEGIN
    INSERT INTO commands_fts(commands_fts, rowid, command) VALUES ('delete', old.id, old.command);
    INSERT INTO commands_fts(rowid, command) VALUES (new.id, new.command);
END;
