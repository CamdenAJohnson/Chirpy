-- +goose Up
CREATE TABLE chirps (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    body TEXT NOT NULL,
    user_id uuid NOT NULL,
    CONSTRAINT kf_user
        FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;