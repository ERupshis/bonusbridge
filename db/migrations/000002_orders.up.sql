--ORDERS
CREATE SCHEMA IF NOT EXISTS orders;

--ORDERS STATUS
CREATE TABLE IF NOT EXISTS orders.statuses
(
    id SMALLSERIAL PRIMARY KEY,
    status VARCHAR(15) NOT NULL UNIQUE
);

INSERT INTO orders.statuses(status)
VALUES ('NEW'),
       ('PROCESSING'),
       ('INVALID'),
       ('PROCESSED'),
       ('UNDEFINED');

--ORDERS ITSELF
CREATE TABLE IF NOT EXISTS orders.orders
(
    id SERIAL PRIMARY KEY,
    num NUMERIC NOT NULL UNIQUE,
    status_id SMALLINT REFERENCES orders.statuses(id) NOT NULL,
    user_id INTEGER REFERENCES users.users(id) NOT NULL ,
    accrual_status SMALLINT,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
