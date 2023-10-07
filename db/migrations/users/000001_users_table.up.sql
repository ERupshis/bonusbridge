CREATE SCHEMA IF NOT EXISTS users;

--ROLES IMPLEMENTATION
CREATE TABLE IF NOT EXISTS users.roles
(
    id   SMALLSERIAL PRIMARY KEY,
    role_id VARCHAR(10) NOT NULL UNIQUE
);

INSERT INTO users.roles(role_id)
VALUES ('admin'),
       ('user');

--USERS IMPLEMENTATION
CREATE TABLE IF NOT EXISTS persons_data.countries
(
    id   SERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(60) NOT NULL,
    role_id SMALLINT NOT NULL
);
