-- +goose Up
CREATE TABLE accounts (
    id uuid PRIMARY KEY,
    atcoder_id text,
    time_zone text NOT NULL DEFAULT 'Asia/Tokyo',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT accounts_atcoder_id_not_blank
        CHECK (atcoder_id IS NULL OR length(btrim(atcoder_id)) > 0),
    CONSTRAINT accounts_time_zone_not_blank
        CHECK (length(btrim(time_zone)) > 0)
);

CREATE TABLE auth_identities (
    issuer text NOT NULL,
    subject text NOT NULL,
    user_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (issuer, subject),
    CONSTRAINT auth_identities_issuer_not_blank CHECK (length(btrim(issuer)) > 0),
    CONSTRAINT auth_identities_subject_not_blank CHECK (length(btrim(subject)) > 0)
);

CREATE INDEX auth_identities_user_id_index ON auth_identities (user_id);

-- +goose Down
DROP TABLE auth_identities;
DROP TABLE accounts;
