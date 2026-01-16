CREATE TYPE status_type AS ENUM (
    'NEW',
    'REGISTERED',
    'PROCESSED',
    'PROCESSING',
    'INVALID'
);

CREATE TABLE orders (
    id VARCHAR PRIMARY KEY,
    status status_type NOT NULL,
	accrual BIGINT NOT NULL,
	user_id BIGINT REFERENCES users(id),
	updated_at TIMESTAMPTZ NOT NULL
);
