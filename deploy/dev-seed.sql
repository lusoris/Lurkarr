-- Dev seed: creates a default admin account if none exists.
-- Credentials: admin / admin123
CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (id, username, password, auth_provider, is_admin)
SELECT gen_random_uuid(), 'admin', crypt('admin123', gen_salt('bf', 12)), 'local', true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');
