INSERT INTO flags (tenant_id, `key`, description, default_enabled)
SELECT id, 'phase8_manual_default_off', 'Phase 8 manual flag with default off', FALSE
FROM tenants
WHERE app_id = 'app-acme'
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  default_enabled = VALUES(default_enabled),
  archived_at = NULL;

INSERT INTO flags (tenant_id, `key`, description, default_enabled)
SELECT id, 'phase8_manual_default_off', 'Phase 8 manual flag with default off', FALSE
FROM tenants
WHERE app_id = 'app-globex'
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  default_enabled = VALUES(default_enabled),
  archived_at = NULL;

INSERT INTO flags (tenant_id, `key`, description, default_enabled)
SELECT id, 'phase8_manual_default_on', 'Phase 8 manual flag with default on', TRUE
FROM tenants
WHERE app_id = 'app-acme'
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  default_enabled = VALUES(default_enabled),
  archived_at = NULL;

INSERT INTO flags (tenant_id, `key`, description, default_enabled)
SELECT id, 'phase8_manual_default_on', 'Phase 8 manual flag with default on', TRUE
FROM tenants
WHERE app_id = 'app-globex'
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  default_enabled = VALUES(default_enabled),
  archived_at = NULL;

INSERT INTO flag_user_overrides (tenant_id, flag_id, user_id, enabled)
SELECT t.id, f.id, 'user_123', TRUE
FROM tenants t
JOIN flags f ON f.tenant_id = t.id AND f.`key` = 'phase8_manual_default_off'
WHERE t.app_id = 'app-acme'
ON DUPLICATE KEY UPDATE
  enabled = VALUES(enabled),
  updated_at = CURRENT_TIMESTAMP;
