-- migrate:up
CREATE TABLE IF NOT EXISTS "device" (
  "id" SERIAL PRIMARY KEY,
  "device_id" TEXT NOT NULL UNIQUE,
  "name" TEXT NOT NULL,
  "device_type" TEXT NOT NULL,
  "status" TEXT NOT NULL DEFAULT 'offline',
  "last_seen" TIMESTAMP,
  "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX "device_device_id_idx" ON "device" ("device_id");
CREATE INDEX "device_status_idx" ON "device" ("status");

-- migrate:down
DROP INDEX IF EXISTS "device_status_idx";
DROP INDEX IF EXISTS "device_device_id_idx";
DROP TABLE IF EXISTS "device";
