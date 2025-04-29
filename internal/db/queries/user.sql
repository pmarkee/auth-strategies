-- name: GetUserInfo :one
SELECT first_name, last_name FROM user_account WHERE id=$1;

-- name: GetPasswordAuth :one
SELECT ua.id, pa.pw_hash, pa.pw_salt
FROM user_account ua
JOIN password_auth pa ON pa.user_id = ua.id
WHERE ua.email=$1;