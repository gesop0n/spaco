-- name: FindUserIDByAuthIdentity :one
SELECT user_id
FROM auth_identities
WHERE issuer = $1 AND subject = $2;

-- name: CreateAccount :exec
INSERT INTO accounts (id, time_zone)
VALUES ($1, $2);

-- name: CreateAuthIdentity :execrows
INSERT INTO auth_identities (issuer, subject, user_id)
VALUES ($1, $2, $3)
ON CONFLICT (issuer, subject) DO NOTHING;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;

-- name: FindAccountByID :one
SELECT id, atcoder_id, time_zone
FROM accounts
WHERE id = $1;

-- name: UpdateAccountProfile :one
UPDATE accounts
SET
    atcoder_id = $2,
    time_zone = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, atcoder_id, time_zone;
