DELETE FROM users
WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'demo')
	AND LOWER(email) IN (
		'admin@example.com',
		'owner@example.com',
		'staff@example.com',
		'guest@example.com'
	);
