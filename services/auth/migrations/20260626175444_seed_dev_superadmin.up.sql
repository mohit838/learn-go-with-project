WITH tenant_row AS (
	INSERT INTO tenants (name, slug)
	VALUES ('Demo Tenant', 'demo')
	ON CONFLICT (slug) DO UPDATE
	SET name = EXCLUDED.name
	RETURNING id
),
role_row AS (
	SELECT id
	FROM roles
	WHERE name = 'superadmin'
)
INSERT INTO users (tenant_id, role_id, username, email, password_hash)
SELECT
	tenant_row.id,
	role_row.id,
	'superadmin',
	'superadmin@example.com',
	crypt('admin123', gen_salt('bf'))
FROM tenant_row, role_row
WHERE NOT EXISTS (
	SELECT 1
	FROM users u
	WHERE u.tenant_id = tenant_row.id
		AND LOWER(u.email) = 'superadmin@example.com'
);

UPDATE users
SET role_id = (SELECT id FROM roles WHERE name = 'superadmin'),
	is_active = TRUE,
	updated_at = NOW()
WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'demo')
	AND LOWER(email) = 'superadmin@example.com';
