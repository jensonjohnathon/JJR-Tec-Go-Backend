-- Add index for username in users table
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);

-- Add index for role_name in roles table
CREATE INDEX IF NOT EXISTS idx_roles_role_name ON roles (role_name);

-- Add index for user_id in user_roles junction table
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles (user_id);

-- Add index for role_id in user_roles junction table
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles (role_id);