-- +goose Up
CREATE TABLE IF NOT EXISTS __sittella_sessions (
	id TEXT PRIMARY KEY,
	expires_at DATETIME NOT NULL,
	data BLOB NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS __sittella_sessions_update
AFTER UPDATE ON __sittella_sessions
FOR EACH ROW
WHEN NEW.updated_at IS NULL OR NEW.updated_at = OLD.updated_at
BEGIN
	UPDATE __sittella_sessions
	SET updated_at = CURRENT_TIMESTAMP
	WHERE id = NEW.id;
END;
-- +goose StatementEnd
