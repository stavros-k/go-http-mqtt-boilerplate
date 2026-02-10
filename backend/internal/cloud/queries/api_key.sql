-- name: GetAPIKey :one
SELECT * FROM "api_key"
WHERE key_hash = $1 AND revoked = false LIMIT 1;

-- name: ListAPIKeysByUser :many
SELECT * FROM "api_key"
WHERE user_id = $1 AND revoked = false
ORDER BY created_at DESC;

-- name: CreateAPIKey :one
INSERT INTO "api_key" (
    key_hash,
    name,
    user_id,
    permissions,
    expires_at,
    created_at,
    updated_at
  )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE "api_key"
SET last_used = $2, updated_at = $3
WHERE key_hash = $1;

-- name: RevokeAPIKey :exec
UPDATE "api_key"
SET revoked = true, updated_at = $2
WHERE key_hash = $1;

-- name: DeleteAPIKey :exec
DELETE FROM "api_key"
WHERE key_hash = $1;
