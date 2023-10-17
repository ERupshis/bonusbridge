CREATE TABLE IF NOT EXISTS bonuses.withdrawals
(
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users.users(id) NOT NULL,
    order_id INTEGER UNIQUE REFERENCES orders.orders(id) NOT NULL,
    bonus_id INTEGER UNIQUE REFERENCES bonuses.bonuses(id) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);