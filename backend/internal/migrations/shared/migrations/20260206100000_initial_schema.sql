-- migrate:up
CREATE TABLE IF NOT EXISTS "some-table" (
  "id" SERIAL PRIMARY KEY,
  "name" TEXT NOT NULL,
  "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "user" (
  "id" SERIAL PRIMARY KEY,
  "name" TEXT NOT NULL,
  "email" TEXT NOT NULL,
  "password" TEXT NOT NULL,
  "last_login" TIMESTAMP,
  "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_user_some_table FOREIGN KEY (id) REFERENCES "some-table" (id) ON DELETE CASCADE
);

CREATE INDEX "user_name_idx" ON "user" ("name");
CREATE INDEX "user_email_idx" ON "user" ("email");

-- migrate:down
DROP INDEX IF EXISTS "user_email_idx";
DROP INDEX IF EXISTS "user_name_idx";
DROP TABLE IF EXISTS "user";
DROP TABLE IF EXISTS "some-table";
