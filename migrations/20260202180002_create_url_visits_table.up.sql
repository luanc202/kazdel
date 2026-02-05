CREATE TABLE IF NOT EXISTS url_visits (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL REFERENCES shortened_urls(id) ON DELETE CASCADE,
    ip_address INET,
    referrer TEXT,
    user_agent TEXT,
    clicked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_visit_url_id ON url_visits(url_id);
