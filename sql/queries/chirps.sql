-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    uuidv7(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;