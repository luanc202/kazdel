CREATE TABLE IF NOT EXISTS shortened_urls (
    id BIGSERIAL PRIMARY KEY,
    short_slug VARCHAR(12) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_short_slug ON shortened_urls(short_slug);
