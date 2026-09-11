-- name: CreateReviewItems :many
-- pgxのexec modeは[]uuid.UUIDを配列として送れないため、UUIDの配列はtextで受け取って変換する。
INSERT INTO review_items (id, user_id, problem_id, registration_note, registered_at, due_on)
SELECT
    unnest(sqlc.arg(ids)::text[])::uuid,
    unnest(sqlc.arg(user_ids)::text[])::uuid,
    unnest(sqlc.arg(problem_ids)::text[]),
    unnest(sqlc.arg(registration_notes)::text[]),
    unnest(sqlc.arg(registered_ats)::timestamptz[]),
    unnest(sqlc.arg(due_ons)::date[])
ON CONFLICT (user_id, problem_id) DO NOTHING
RETURNING *;

-- name: FindReviewItemsByProblemIDs :many
SELECT *
FROM review_items
WHERE user_id = sqlc.arg(user_id)
  AND problem_id = ANY(sqlc.arg(problem_ids)::text[]);

-- name: FindReviewItemByID :one
SELECT *
FROM review_items
WHERE user_id = $1 AND id = $2;

-- name: LockReviewItemByID :one
SELECT *
FROM review_items
WHERE user_id = $1 AND id = $2
FOR UPDATE;

-- name: UpdateReviewItemSchedule :one
UPDATE review_items
SET
    state = sqlc.arg(state),
    due_on = sqlc.arg(due_on),
    interval_days = sqlc.arg(interval_days),
    stability = sqlc.narg(stability),
    difficulty = sqlc.narg(difficulty),
    reps = sqlc.arg(reps),
    lapses = sqlc.arg(lapses),
    last_reviewed_at = sqlc.narg(last_reviewed_at),
    updated_at = now()
WHERE user_id = sqlc.arg(user_id) AND id = sqlc.arg(id)
RETURNING *;

-- name: UpdateReviewItemPausedAt :one
UPDATE review_items
SET
    paused_at = sqlc.narg(paused_at),
    updated_at = now()
WHERE user_id = sqlc.arg(user_id) AND id = sqlc.arg(id)
RETURNING *;

-- name: ListDueReviewItems :many
SELECT *
FROM review_items
WHERE user_id = sqlc.arg(user_id)
  AND paused_at IS NULL
  AND due_on <= sqlc.arg(today)
ORDER BY due_on, registered_at, id;

-- name: ListUpcomingReviewItems :many
SELECT *
FROM review_items
WHERE user_id = sqlc.arg(user_id)
  AND paused_at IS NULL
  AND due_on > sqlc.arg(today)
ORDER BY due_on, registered_at, id
LIMIT sqlc.arg(max_items);

-- name: ListReviewItems :many
SELECT *
FROM review_items
WHERE user_id = $1
ORDER BY registered_at DESC, id DESC;

-- name: ListRegisteredProblemIDs :many
SELECT problem_id
FROM review_items
WHERE user_id = sqlc.arg(user_id)
  AND problem_id = ANY(sqlc.arg(problem_ids)::text[])
ORDER BY problem_id;

-- name: CreateReviewLog :one
INSERT INTO review_logs (
    id,
    review_item_id,
    user_id,
    request_id,
    result,
    rating,
    performed_at,
    note,
    state_before,
    elapsed_days,
    last_interval_days,
    interval_days,
    stability,
    difficulty,
    next_due_on
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(review_item_id),
    sqlc.arg(user_id),
    sqlc.arg(request_id),
    sqlc.arg(result),
    sqlc.arg(rating),
    sqlc.arg(performed_at),
    sqlc.arg(note),
    sqlc.arg(state_before),
    sqlc.arg(elapsed_days),
    sqlc.arg(last_interval_days),
    sqlc.arg(interval_days),
    sqlc.arg(stability),
    sqlc.arg(difficulty),
    sqlc.arg(next_due_on)
)
RETURNING *;

-- name: FindReviewLogByRequestID :one
SELECT *
FROM review_logs
WHERE user_id = $1 AND request_id = $2;

-- name: FindLatestReviewLogs :many
SELECT DISTINCT ON (review_item_id) *
FROM review_logs
WHERE user_id = sqlc.arg(user_id)
  AND review_item_id = ANY(sqlc.arg(review_item_ids)::text[]::uuid[])
ORDER BY review_item_id, performed_at DESC, created_at DESC;

-- name: ListReviewLogs :many
SELECT *
FROM review_logs
WHERE user_id = $1
ORDER BY performed_at, created_at;

-- name: CountReviewLogsPerformedBetween :one
SELECT count(*)
FROM review_logs
WHERE user_id = sqlc.arg(user_id)
  AND performed_at >= sqlc.arg(performed_from)
  AND performed_at < sqlc.arg(performed_until);
