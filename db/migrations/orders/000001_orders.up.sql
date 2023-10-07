CREATE SCHEMA IF NOT EXISTS orders;

--ORDERS STATUS
CREATE TABLE IF NOT EXISTS orders.statuses
(
    id SMALLSERIAL PRIMARY KEY,
    status_id VARCHAR(15) NOT NULL UNIQUE
);

INSERT INTO orders.statuses(status_id)
VALUES ('NEW'),
       ('PROCESSING'),
       ('INVALID'),
       ('PROCESSED');

--ORDERS ITSELF
CREATE TABLE IF NOT EXISTS persons_data.countries
(
    id SERIAL PRIMARY KEY,
    num NUMERIC NOT NULL UNIQUE,
    status_id SMALLINT REFERENCES orders.statuses(id) NOT NULL,
    user_id INTEGER NOT NULL  users.users(id),
    accrual_status SMALLINT,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);