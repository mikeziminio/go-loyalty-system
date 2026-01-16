CREATE TABLE withdrawals (
    id BIGSERIAL PRIMARY KEY,
    sum BIGINT NOT NULL,
    order_id VARCHAR NOT NULL,
	user_id BIGINT REFERENCES users(id),
	processed_at TIMESTAMPTZ NOT NULL
);
