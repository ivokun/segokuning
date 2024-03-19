CREATE TABLE IF NOT EXISTS friends(
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  friend_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT (TIMEZONE('UTC', NOW())),
  modified_at TIMESTAMP DEFAULT (TIMEZONE('UTC', NOW())),
);

-- Trigger to update modified_at column
CREATE OR REPLACE FUNCTION update_friends_modified_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.modified_at = TIMEZONE('UTC', NOW());
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_friends_modified_at_column_trigger
BEFORE UPDATE ON friends
FOR EACH ROW
EXECUTE FUNCTION update_friends_modified_at_column();

-- Trigger to add another row with friend_id as user_id and user_id as friend_id
CREATE OR REPLACE FUNCTION add_friend_connection()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO friends (id, user_id, friend_id)
  VALUES (NEW.id, NEW.friend_id, NEW.user_id);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER add_friend_connection_trigger
AFTER INSERT ON friends 
FOR EACH row
EXECUTE FUNCTION add_friend_connection();

