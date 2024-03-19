CREATE TABLE IF NOT EXISTS posts(
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  post_in_html VARCHAR (50) NOT NULL,
  created_at TIMESTAMP DEFAULT (TIMEZONE('UTC', NOW()))
);

-- Add post tags table
CREATE TABLE IF NOT EXISTS post_tags(
  post_id TEXT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  tags TEXT[] NOT NULL DEFAULT '{}'
);

CREATE INDEX post_tags_post_id_index ON post_tags USING GIN (tags);
