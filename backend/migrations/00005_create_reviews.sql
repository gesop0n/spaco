-- +goose Up
-- 復習対象。Ankiのcardに相当し、FSRSのスケジューラ状態を持つ。review moduleだけが更新する。
-- user_idとproblem_idは他moduleのデータを指すが、moduleの境界を保つため外部キーにしない。
CREATE TABLE review_items (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    problem_id text NOT NULL,
    registration_note text NOT NULL DEFAULT '',
    registered_at timestamptz NOT NULL,
    -- NULLでなければ一時停止中。
    paused_at timestamptz,

    -- new: 登録後に一度も再挑戦していない。review: FSRSの記憶状態を持つ。
    state text NOT NULL DEFAULT 'new',
    -- ユーザーのタイムゾーンでの次回予定日。
    due_on date NOT NULL,
    -- 直前の記録で決めた間隔。Ankiのscheduled_daysに相当する。
    interval_days integer NOT NULL DEFAULT 0,
    stability double precision,
    difficulty double precision,
    reps integer NOT NULL DEFAULT 0,
    lapses integer NOT NULL DEFAULT 0,
    last_reviewed_at timestamptz,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT review_items_user_problem_unique UNIQUE (user_id, problem_id),
    CONSTRAINT review_items_registration_note_length
        CHECK (char_length(registration_note) <= 500),
    CONSTRAINT review_items_state_valid CHECK (
        (
            state = 'new'
            AND stability IS NULL
            AND difficulty IS NULL
            AND last_reviewed_at IS NULL
            AND reps = 0
        )
        OR (
            state = 'review'
            AND stability > 0
            AND difficulty BETWEEN 1 AND 10
            AND last_reviewed_at IS NOT NULL
            AND reps > 0
        )
    ),
    CONSTRAINT review_items_counts_non_negative
        CHECK (interval_days >= 0 AND reps >= 0 AND lapses >= 0)
);

-- 今日の復習は、一時停止していない復習対象を予定日で絞り込む。
CREATE INDEX review_items_due_index ON review_items (user_id, due_on) WHERE paused_at IS NULL;

-- 再挑戦の履歴。Ankiのrevlogに相当し、登録は含めない。
CREATE TABLE review_logs (
    id uuid PRIMARY KEY,
    review_item_id uuid NOT NULL REFERENCES review_items (id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    -- 保存ボタンの二重送信を判定するため、入力画面ごとにクライアントが生成するID。
    request_id uuid NOT NULL,
    -- independent: 自力でAC。assisted: 解説・ヒントを見てAC。unsolved: ACできなかった。
    result text NOT NULL,
    -- FSRSへの入力。1: Again、2: Hard、3: Good、4: Easy。
    rating smallint NOT NULL,
    performed_at timestamptz NOT NULL,
    note text NOT NULL DEFAULT '',

    -- 記録時点のスケジューラの入出力。FSRSのパラメータ最適化や計算の検証に使う。
    state_before text NOT NULL,
    elapsed_days integer NOT NULL,
    last_interval_days integer NOT NULL,
    interval_days integer NOT NULL,
    stability double precision NOT NULL,
    difficulty double precision NOT NULL,
    next_due_on date NOT NULL,

    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT review_logs_request_unique UNIQUE (user_id, request_id),
    CONSTRAINT review_logs_result_valid
        CHECK (result IN ('independent', 'assisted', 'unsolved')),
    CONSTRAINT review_logs_rating_valid CHECK (rating BETWEEN 1 AND 4),
    CONSTRAINT review_logs_note_length CHECK (char_length(note) <= 1000),
    CONSTRAINT review_logs_state_before_valid CHECK (state_before IN ('new', 'review')),
    CONSTRAINT review_logs_days_valid
        CHECK (elapsed_days >= 0 AND last_interval_days >= 0 AND interval_days >= 1)
);

CREATE INDEX review_logs_item_index ON review_logs (review_item_id, performed_at);
CREATE INDEX review_logs_user_performed_index ON review_logs (user_id, performed_at);

-- +goose Down
DROP TABLE review_logs;
DROP TABLE review_items;
