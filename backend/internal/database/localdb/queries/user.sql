-- name: CreateUser :one
INSERT INTO "user" (
    name,
    email,
    password,
    created_at,
    updated_at
  )
VALUES (
    :name,
    :email,
    :password,
    :created_at,
    :updated_at
  )
RETURNING *;
-- name: CreateUserWithPassword :one
INSERT INTO "user" (
    name,
    email,
    password,
    created_at,
    updated_at
  )
VALUES (
    :name,
    :email,
    :password,
    :created_at,
    :updated_at
  )
RETURNING *;
