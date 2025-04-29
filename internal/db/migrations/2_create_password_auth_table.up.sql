CREATE TABLE IF NOT EXISTS password_auth (
     id SERIAL PRIMARY KEY,
     user_id UUID NOT NULL UNIQUE REFERENCES user_account(id) ON DELETE CASCADE,
     pw_hash BYTEA NOT NULL,
     pw_salt BYTEA NOT NULL
);