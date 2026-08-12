ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

CREATE TABLE app_settings (
    key VARCHAR(255) PRIMARY KEY,
    value VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO app_settings (key, value) VALUES ('signup_enabled', 'true');
INSERT INTO app_settings (key, value) VALUES ('email_validation_enabled', 'false');
