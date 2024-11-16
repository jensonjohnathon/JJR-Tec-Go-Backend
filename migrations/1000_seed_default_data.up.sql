-- Enable pgcrypto extension for password encryption
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Make sure to update this password manually or in migrations if it changes
INSERT INTO users (username, email, password)
VALUES ('admin', 'admin@example.com', crypt('DefaultAdminPassword', gen_salt('bf', 12)))
ON CONFLICT DO NOTHING;

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE u.username = 'admin' AND r.role_name = 'admin'
ON CONFLICT DO NOTHING;