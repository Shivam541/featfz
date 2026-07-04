CREATE TABLE IF NOT EXISTS flags (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL,
  `key` VARCHAR(255) NOT NULL,
  description TEXT NULL,
  default_enabled BOOLEAN NOT NULL,
  archived_at TIMESTAMP NULL DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_flags_tenant_key (tenant_id, `key`),
  KEY idx_flags_tenant_archived (tenant_id, archived_at),
  CONSTRAINT fk_flags_tenant_id
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
    ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS flag_user_overrides (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL,
  flag_id BIGINT UNSIGNED NOT NULL,
  user_id VARCHAR(255) NOT NULL,
  enabled BOOLEAN NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_flag_user_overrides_tenant_flag_user (tenant_id, flag_id, user_id),
  KEY idx_flag_user_overrides_tenant_flag_user (tenant_id, flag_id, user_id),
  CONSTRAINT fk_flag_user_overrides_tenant_id
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
    ON DELETE CASCADE,
  CONSTRAINT fk_flag_user_overrides_flag_id
    FOREIGN KEY (flag_id) REFERENCES flags(id)
    ON DELETE CASCADE
);
