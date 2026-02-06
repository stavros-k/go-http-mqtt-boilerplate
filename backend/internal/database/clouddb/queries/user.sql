-- name: CreateUser :one
INSERT INTO "user" (
    name,
    email,
    password,
    created_at,
    updated_at
  )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
-- name: CreateUserWithPassword :one
INSERT INTO "user" (
    name,
    email,
    password,
    created_at,
    updated_at
  )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
