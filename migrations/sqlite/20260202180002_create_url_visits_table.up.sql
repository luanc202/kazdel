CREATE TABLE IF NOT EXISTS url_visits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url_id INTEGER REFERENCES shortened_urls(id) ON DELETE CASCADE,
    visited_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    referrer TEXT
);
