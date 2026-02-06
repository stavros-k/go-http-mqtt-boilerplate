-- migrate:up
ALTER TABLE "user"
ADD COLUMN "last_login" TIMESTAMP;
-- migrate:down
ALTER TABLE "user" DROP COLUMN "last_login";
