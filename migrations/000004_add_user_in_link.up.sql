ALTER TABLE links
ADD COLUMN userID uuid;

CREATE INDEX idx_links_user_id ON links(userID);
