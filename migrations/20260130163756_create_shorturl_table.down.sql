-- Down migration
-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS url_visits;
DROP TABLE IF EXISTS shortened_urls;
DROP TABLE IF EXISTS users;

-- Optional: Drop indexes (though dropping tables usually handles this)
DROP INDEX IF EXISTS idx_visit_url_id;
DROP INDEX IF EXISTS idx_short_slug;