-- Up migration
-- 1. Create Users Table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create ShortenedURLs Table
CREATE TABLE shortened_urls (
    id BIGSERIAL PRIMARY KEY,
    short_slug VARCHAR(12) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Index the slug for lightning-fast lookups
CREATE INDEX idx_short_slug ON shortened_urls(short_slug);

-- 3. Create URLVisits Table (Analytics)
CREATE TABLE url_visits (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL REFERENCES shortened_urls(id) ON DELETE CASCADE,
    ip_address INET, -- Using INET type for optimized IP storage
    referrer TEXT,
    user_agent TEXT,
    clicked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index url_id for faster analytics queries
CREATE INDEX idx_visit_url_id ON url_visits(url_id);