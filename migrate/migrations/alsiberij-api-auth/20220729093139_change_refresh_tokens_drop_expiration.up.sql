ALTER TABLE refresh_tokens DROP COLUMN "expiresAt";

ALTER TABLE refresh_tokens ALTER COLUMN "lastUsedAt" SET DEFAULT CURRENT_TIMESTAMP;
UPDATE refresh_tokens SET "lastUsedAt" = CURRENT_TIMESTAMP WHERE "lastUsedAt" IS NULL;
ALTER TABLE refresh_tokens ALTER COLUMN "lastUsedAt" SET NOT NULL;