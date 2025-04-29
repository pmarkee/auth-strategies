-- name: GetUserInfo :one
SELECT first_name, last_name FROM user_account WHERE id=$1;