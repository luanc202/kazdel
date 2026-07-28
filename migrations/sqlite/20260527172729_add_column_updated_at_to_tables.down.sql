ALTER TABLE shortened_urls DROP COLUMN updated_at;
ALTER TABLE url_visits DROP COLUMN updated_at;
ALTER TABLE url_visits DROP COLUMN created_at;
ALTER TABLE users DROP COLUMN deleted_at;
