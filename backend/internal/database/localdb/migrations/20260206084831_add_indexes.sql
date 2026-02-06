-- migrate:up
CREATE INDEX IF NOT EXISTS "user_name_idx" ON "user" ("name");
CREATE INDEX IF NOT EXISTS "user_email_idx" ON "user" ("email");
-- migrate:down
DROP INDEX IF EXISTS "user_name_idx";
DROP INDEX IF EXISTS "user_email_idx";
