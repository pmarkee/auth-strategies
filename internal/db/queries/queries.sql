-- name: GetUserInfo :one
SELECT first_name, last_name FROM user_account WHERE id=$1;

-- name: GetPasswordAuth :one
SELECT ua.id, pa.pw_hash, pa.pw_salt
FROM user_account ua
JOIN password_auth pa ON pa.user_id = ua.id
WHERE ua.email=$1;

-- name: EmailTaken :one
SELECT
    CASE WHEN EXISTS (
        SELECT 1
        FROM user_account
        WHERE email = $1
    ) THEN true ELSE false END;

-- name: CreateUser :one
INSERT INTO user_account (email, first_name, last_name) VALUES ($1, $2, $3) RETURNING id;

-- name: CreatePasswordAuth :exec
INSERT INTO password_auth (user_id, pw_hash, pw_salt) VALUES ($1, $2, $3);

-- name: ApiKeyPublicIdTaken :one
SELECT
    CASE WHEN EXISTS (
        SELECT 1 FROM api_key WHERE public_id=$1
    ) THEN true ELSE false END;

-- name: CreateApiKey :exec
INSERT INTO api_key (user_id, public_id, secret_hash, secret_salt)
VALUES ($1, $2, $3, $4);

-- name: FindApiKey :one
SELECT user_id, public_id, secret_hash, secret_salt
FROM api_key
WHERE public_id=$1;