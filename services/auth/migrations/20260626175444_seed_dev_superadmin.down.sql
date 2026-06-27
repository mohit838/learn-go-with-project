DELETE FROM users
WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'demo')
	AND LOWER(email) = 'superadmin@example.com';

DELETE FROM tenants
WHERE slug = 'demo'
	AND NOT EXISTS (
		SELECT 1
		FROM users
		WHERE users.tenant_id = tenants.id
	);
