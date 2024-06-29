CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            passport_number VARCHAR(20) NOT NULL,
            surname VARCHAR(100),
            name VARCHAR(100),
            patronymic VARCHAR(100),
            address TEXT,
            created_at TIMESTAMPTZ,
            updated_at TIMESTAMPTZ
);

CREATE TABLE tasks (
            id SERIAL PRIMARY KEY,
            user_id INTEGER REFERENCES users(id),
            name VARCHAR(100),
            hours INTEGER,
            minutes INTEGER,
            created_at TIMESTAMPTZ,
            updated_at TIMESTAMPTZ,
            start_time TIMESTAMPTZ,
            end_time TIMESTAMPTZ
);
