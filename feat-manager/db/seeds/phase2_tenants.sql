INSERT INTO tenants (name, app_id, jwt_secret)
VALUES
  ('acme', 'app-acme', 'acme-secret'),
  ('globex', 'app-globex', 'globex-secret')
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  jwt_secret = VALUES(jwt_secret);
