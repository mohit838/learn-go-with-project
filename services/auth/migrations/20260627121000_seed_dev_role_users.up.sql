WITH tenant_row AS (
	SELECT id
	FROM tenants
	WHERE slug = 'demo'
),
seed_users AS (
	SELECT 'admin' AS role_name, 'admin' AS username, 'admin@example.com' AS email
	UNION ALL
	SELECT 'owner', 'owner', 'owner@example.com'
	UNION ALL
	SELECT 'staff', 'staff', 'staff@example.com'
	UNION ALL
	SELECT 'guest', 'guest', 'guest@example.com'
)
INSERT INTO users (tenant_id, role_id, username, email, password_hash)
SELECT
	tenant_row.id,
	roles.id,
	seed_users.username,
	seed_users.email,
	crypt('admin123', gen_salt('bf'))
FROM seed_users
JOIN roles ON roles.name = seed_users.role_name
CROSS JOIN tenant_row
WHERE NOT EXISTS (
	SELECT 1
	FROM users u
	WHERE u.tenant_id = tenant_row.id
		AND LOWER(u.email) = seed_users.email
);

UPDATE users
SET role_id = roles.id,
	is_active = TRUE,
	updated_at = NOW()
FROM roles
WHERE users.tenant_id = (SELECT id FROM tenants WHERE slug = 'demo')
	AND LOWER(users.email) = roles.name || '@example.com'
	AND roles.name IN ('admin', 'owner', 'staff', 'guest');
