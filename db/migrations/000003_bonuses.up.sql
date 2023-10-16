--BONUSES
CREATE SCHEMA IF NOT EXISTS bonuses;

--BALANCES
CREATE TABLE IF NOT EXISTS bonuses.bonuses
(
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE REFERENCES users.users(id) NOT NULL,
    --add order_id
    balance NUMERIC(9,2) DEFAULT 0, --fix. need table of positive and negative incomings.
    withdrawn NUMERIC(9,2) DEFAULT 0 -- remove
);

CREATE TABLE IF NOT EXISTS bonuses.withdrawals
(
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users.users(id) NOT NULL,
    order_num NUMERIC NOT NULL,
    sum NUMERIC(9,2) DEFAULT 0,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);