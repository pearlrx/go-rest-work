CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    passport_number VARCHAR(20) NOT NULL,
    surname VARCHAR(50) NOT NULL,
    name VARCHAR(50) NOT NULL,
    patronymic VARCHAR(50),
    address VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);