ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO app_settings (key, value) VALUES ('signup_enabled', 'true');
INSERT INTO app_settings (key, value) VALUES ('email_validation_enabled', 'false');
