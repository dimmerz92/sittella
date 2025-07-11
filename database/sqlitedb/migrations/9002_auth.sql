-- +goose Up
CREATE TABLE IF NOT EXISTS __sittella_auth (
	id TEXT PRIMARY KEY DEFAULT (uuid()),
	expires_at DATETIME NOT NULL,
	data BLOB NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS __sittella_auth_update
AFTER UPDATE ON __sittella_auth
FOR EACH ROW
WHEN NEW.updated_at IS NULL OR NEW.updated_at = OLD.updated_at
BEGIN
	UPDATE __sittella_auth
	SET updated_at = CURRENT_TIMESTAMP
	WHERE id = NEW.id;
END;
-- +goose StatementEnd
