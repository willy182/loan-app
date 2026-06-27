-- +goose Up
CREATE TABLE IF NOT EXISTS loans (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     VARCHAR(255) NOT NULL,
    mrp         BIGINT NOT NULL,
    dp          BIGINT NOT NULL,
    vehicle_year INTEGER NOT NULL,
    police_number  VARCHAR(50) NOT NULL,
    machine_number VARCHAR(255) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'submitted',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, police_number)
);

-- +goose Down
DROP TABLE IF EXISTS loans;
