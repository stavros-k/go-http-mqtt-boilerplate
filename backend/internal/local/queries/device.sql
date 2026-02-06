-- name: GetDevice :one
SELECT * FROM "device"
WHERE device_id = $1 LIMIT 1;

-- name: ListDevices :many
SELECT * FROM "device"
ORDER BY created_at DESC;

-- name: CreateDevice :one
INSERT INTO "device" (
    device_id,
    name,
    device_type,
    status,
    created_at,
    updated_at
  )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateDeviceStatus :exec
UPDATE "device"
SET status = $2, last_seen = $3, updated_at = $4
WHERE device_id = $1;

-- name: DeleteDevice :exec
DELETE FROM "device"
WHERE device_id = $1;
