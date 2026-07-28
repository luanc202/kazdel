ALTER TABLE shortened_urls 
ADD COLUMN description TEXT,
ADD COLUMN password_hash VARCHAR(255);
