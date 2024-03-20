CREATE TABLE IF NOT EXISTS users(
  id TEXT PRIMARY KEY,
  name VARCHAR (50) NOT NULL,
  email VARCHAR (50) NOT NULL,
  phone VARCHAR (50) NOT NULL,
  image_url TEXT NOT NULL,
  password VARCHAR (255) NOT NULL,
  created_at TIMESTAMP DEFAULT (TIMEZONE('UTC', NOW())),
  modified_at TIMESTAMP DEFAULT (TIMEZONE('UTC', NOW())),

  UNIQUE (email, phone)
);

-- Trigger to update modified_at column
CREATE OR REPLACE FUNCTION update_users_modified_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.modified_at = TIMEZONE('UTC', NOW());
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_modified_at_column_trigger
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_users_modified_at_column();

