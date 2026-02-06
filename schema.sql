CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE IF NOT EXISTS "some-table" (
  "id" INTEGER PRIMARY KEY,
  "name" TEXT NOT NULL,
  "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS "user" (
  "id" INTEGER PRIMARY KEY,
  "name" TEXT NOT NULL,
  "email" TEXT NOT NULL,
  "password" TEXT NOT NULL,
  "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, "last_login" TIMESTAMP,
  FOREIGN KEY (id) REFERENCES "some-table" (id) ON DELETE CASCADE
);
CREATE INDEX "user_name_idx" ON "user" ("name");
CREATE INDEX "user_email_idx" ON "user" ("email");
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20251009092116'),
  ('20251009104248'),
  ('20260206084831');
