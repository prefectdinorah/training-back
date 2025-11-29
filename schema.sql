-- Create users table with UUID primary key
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  login VARCHAR(50) UNIQUE NOT NULL,
  name VARCHAR(100),
  surname VARCHAR(100),
  email VARCHAR(255) UNIQUE NOT NULL,
  phone VARCHAR(20),
  trainer UUID REFERENCES users(id) ON DELETE SET NULL,
  is_trainer BOOLEAN DEFAULT false,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_login ON users(login);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
