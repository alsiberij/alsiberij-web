ALTER TABLE refresh_tokens ADD COLUMN "isExpired" BOOLEAN DEFAULT FALSE NOT NULL;