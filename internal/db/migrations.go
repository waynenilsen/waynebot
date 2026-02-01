package db

import (
	"database/sql"
	"fmt"
)

// migration represents a single versioned schema migration.
type migration struct {
	Version int
	SQL     string
}

var migrations = []migration{
	{
		Version: 1,
		SQL: `
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);

CREATE TABLE ws_tickets (
    ticket TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);

CREATE TABLE invites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT UNIQUE NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    used_by INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id INTEGER NOT NULL REFERENCES channels(id),
    author_id INTEGER NOT NULL,
    author_type TEXT NOT NULL CHECK(author_type IN ('human', 'agent')),
    author_name TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_messages_channel_id ON messages(channel_id, id);

CREATE TABLE personas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    system_prompt TEXT NOT NULL,
    model TEXT NOT NULL DEFAULT 'anthropic/claude-sonnet-4-20250514',
    tools_enabled TEXT NOT NULL DEFAULT '["shell_exec","file_read","file_write","http_fetch"]',
    temperature REAL NOT NULL DEFAULT 0.7,
    max_tokens INTEGER NOT NULL DEFAULT 4096,
    cooldown_secs INTEGER NOT NULL DEFAULT 30,
    max_tokens_per_hour INTEGER NOT NULL DEFAULT 100000,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE persona_channels (
    persona_id INTEGER NOT NULL REFERENCES personas(id) ON DELETE CASCADE,
    channel_id INTEGER NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    PRIMARY KEY (persona_id, channel_id)
);

CREATE TABLE actor_cursors (
    persona_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    last_seen_message_id INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(persona_id, channel_id)
);

CREATE TABLE llm_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    persona_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    model TEXT NOT NULL,
    messages_json TEXT NOT NULL,
    response_json TEXT NOT NULL,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_llm_calls_persona_time ON llm_calls(persona_id, created_at);

CREATE TABLE tool_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    persona_id INTEGER NOT NULL,
    tool_name TEXT NOT NULL,
    args_json TEXT NOT NULL,
    output_text TEXT,
    error_text TEXT,
    duration_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`,
	},
	{
		Version: 2,
		SQL: `
-- Widen author_type CHECK to include 'connector'.
CREATE TABLE messages_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id INTEGER NOT NULL REFERENCES channels(id),
    author_id INTEGER NOT NULL,
    author_type TEXT NOT NULL CHECK(author_type IN ('human', 'agent', 'connector')),
    author_name TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO messages_new SELECT * FROM messages;
DROP TABLE messages;
ALTER TABLE messages_new RENAME TO messages;
CREATE INDEX idx_messages_channel_id ON messages(channel_id, id);
`,
	},
	{
		Version: 3,
		SQL: `
CREATE TABLE read_positions (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_id INTEGER NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    last_read_message_id INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, channel_id)
);
`,
	},
	{
		Version: 4,
		SQL: `
CREATE TABLE reactions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id  INTEGER NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    author_id   INTEGER NOT NULL,
    author_type TEXT    NOT NULL CHECK(author_type IN ('human', 'agent', 'connector')),
    emoji       TEXT    NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(message_id, author_id, author_type, emoji)
);
CREATE INDEX idx_reactions_message_id ON reactions(message_id);
`,
	},
	{
		Version: 5,
		SQL: `
ALTER TABLE channels ADD COLUMN is_dm BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE channels ADD COLUMN created_by INTEGER;

CREATE TABLE dm_participants (
    channel_id INTEGER NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id    INTEGER,
    persona_id INTEGER,
    UNIQUE (channel_id, user_id, persona_id)
);
CREATE INDEX idx_dm_participants_user ON dm_participants(user_id);
CREATE INDEX idx_dm_participants_persona ON dm_participants(persona_id);
`,
	},
	{
		Version: 6,
		SQL: `
CREATE TABLE channel_members (
    channel_id INTEGER NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('owner', 'member')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (channel_id, user_id)
);
CREATE INDEX idx_channel_members_user ON channel_members(user_id);
`,
	},
	{
		Version: 7,
		SQL: `
-- Backfill: add all existing users as owners of all existing non-DM channels
-- that have no members yet. This prevents channels from disappearing after
-- membership-based visibility filtering was introduced.
INSERT OR IGNORE INTO channel_members (channel_id, user_id, role)
SELECT c.id, u.id, 'owner'
FROM channels c
CROSS JOIN users u
WHERE c.is_dm = 0
  AND NOT EXISTS (
    SELECT 1 FROM channel_members cm WHERE cm.channel_id = c.id
  );
`,
	},
	{
		Version: 8,
		SQL: `
CREATE TABLE projects (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    path        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`,
	},
	{
		Version: 9,
		SQL: `
CREATE TABLE channel_projects (
    channel_id INTEGER NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    PRIMARY KEY (channel_id, project_id)
);
`,
	},
}

// migrate runs all pending migrations inside a transaction.
func (d *DB) migrate() error {
	d.writeMu.Lock()
	defer d.writeMu.Unlock()

	// Ensure schema_version table exists.
	if _, err := d.SQL.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}

	for _, m := range migrations {
		applied, err := d.isApplied(m.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		tx, err := d.SQL.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for v%d: %w", m.Version, err)
		}

		if _, err := tx.Exec(m.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply v%d: %w", m.Version, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (?)", m.Version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record v%d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit v%d: %w", m.Version, err)
		}
	}

	return nil
}

func (d *DB) isApplied(version int) (bool, error) {
	var count int
	err := d.SQL.QueryRow("SELECT COUNT(*) FROM schema_version WHERE version = ?", version).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return count > 0, nil
}

// SchemaVersion returns the highest applied migration version.
func (d *DB) SchemaVersion() (int, error) {
	var v int
	err := d.SQL.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&v)
	return v, err
}
