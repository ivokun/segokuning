CREATE TABLE IF NOT EXISTS friends(
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  friend_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT (TIMEZONE('UTC', NOW())),

  PRIMARY KEY (user_id, friend_id)
);

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

