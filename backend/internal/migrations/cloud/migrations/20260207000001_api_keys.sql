-- migrate:up
CREATE TABLE IF NOT EXISTS "api_key" (
  "id" SERIAL PRIMARY KEY,
  "key_hash" TEXT NOT NULL UNIQUE,
  "name" TEXT NOT NULL,
  "user_id" INTEGER NOT NULL,
  "permissions" JSONB NOT NULL DEFAULT '[]',
  "last_used" TIMESTAMP,
  "expires_at" TIMESTAMP,
  "revoked" BOOLEAN NOT NULL DEFAULT false,
  "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_api_key_user FOREIGN KEY (user_id) REFERENCES "user" (id) ON DELETE CASCADE
);

CREATE INDEX "api_key_key_hash_idx" ON "api_key" ("key_hash");
CREATE INDEX "api_key_user_id_idx" ON "api_key" ("user_id");
CREATE INDEX "api_key_revoked_idx" ON "api_key" ("revoked");

-- migrate:down
DROP INDEX IF EXISTS "api_key_revoked_idx";
DROP INDEX IF EXISTS "api_key_user_id_idx";
DROP INDEX IF EXISTS "api_key_key_hash_idx";
DROP TABLE IF EXISTS "api_key";
