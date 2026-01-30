-- name: CreateUser :one
INSERT INTO "user" (
    name,
    email,
    password,
    created_at,
    updated_at,
    last_login
  )
VALUES (
    :name,
    :email,
    :password,
    :created_at,
    :updated_at,
    :last_login
  )
RETURNING *;
-- name: CreateUserWithPassword :one
INSERT INTO "user" (
    name,
    email,
    password,
    created_at,
    updated_at,
    last_login
  )
VALUES (
    :name,
    :email,
    :password,
    :created_at,
    :updated_at,
    :last_login
  )
RETURNING *;
