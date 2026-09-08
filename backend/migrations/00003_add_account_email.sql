-- +goose Up
-- 既存Accountは、同じSupabase projectのauth.usersからメールアドレスを補完する。
ALTER TABLE accounts ADD COLUMN email text;

-- +goose StatementBegin
DO $$
BEGIN
    IF to_regclass('auth.users') IS NOT NULL THEN
        EXECUTE $backfill$
            UPDATE accounts AS account
            SET email = auth_user.email
            FROM auth_identities AS identity
            JOIN auth.users AS auth_user ON auth_user.id::text = identity.subject
            WHERE identity.user_id = account.id
              AND auth_user.email IS NOT NULL
        $backfill$;
    END IF;

    IF EXISTS (SELECT 1 FROM accounts WHERE email IS NULL) THEN
        RAISE EXCEPTION 'account email could not be backfilled from auth.users';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE accounts
    ALTER COLUMN email SET NOT NULL,
    ADD CONSTRAINT accounts_email_not_blank CHECK (length(btrim(email)) > 0);

-- +goose Down
ALTER TABLE accounts DROP COLUMN email;
