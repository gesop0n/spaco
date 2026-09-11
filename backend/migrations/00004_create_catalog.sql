-- +goose Up
-- AtCoder Problemsから取り込むコンテスト。catalog moduleだけが更新する。
CREATE TABLE contests (
    id text PRIMARY KEY,
    title text NOT NULL,
    start_at timestamptz NOT NULL,
    -- 常設コンテストの期間は約100年で、integerの上限を超えるためbigintで持つ。
    duration_seconds bigint NOT NULL,
    rate_change text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT contests_id_not_blank CHECK (length(btrim(id)) > 0),
    CONSTRAINT contests_duration_non_negative CHECK (duration_seconds >= 0)
);

CREATE INDEX contests_start_at_index ON contests (start_at DESC);

-- 問題URLから登録した問題は、AtCoder Problemsに情報がなくても行を作り、同期時に補完する。
CREATE TABLE problems (
    id text PRIMARY KEY,
    -- 表示と問題URLに使うコンテスト。未同期のコンテストも入るため、外部キーにしない。
    contest_id text NOT NULL,
    problem_index text,
    name text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT problems_id_not_blank CHECK (length(btrim(id)) > 0),
    CONSTRAINT problems_contest_id_not_blank CHECK (length(btrim(contest_id)) > 0)
);

-- ABCとARCの同時開催などで、1つの問題が複数のコンテストに属する。
CREATE TABLE contest_problems (
    contest_id text NOT NULL REFERENCES contests (id) ON DELETE CASCADE,
    problem_id text NOT NULL REFERENCES problems (id) ON DELETE CASCADE,
    problem_index text NOT NULL,
    PRIMARY KEY (contest_id, problem_id)
);

CREATE INDEX contest_problems_problem_id_index ON contest_problems (problem_id);

-- +goose Down
DROP TABLE contest_problems;
DROP TABLE problems;
DROP TABLE contests;
