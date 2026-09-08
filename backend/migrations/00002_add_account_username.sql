-- +goose Up
-- 初期設定前のAccountと既存Accountは、ユーザー名を未設定のまま保持する。
ALTER TABLE accounts
    ADD COLUMN username text,
    ADD CONSTRAINT accounts_username_length
        CHECK (username IS NULL OR char_length(btrim(username)) BETWEEN 1 AND 40);

-- +goose Down
ALTER TABLE accounts DROP COLUMN username;
