-- name: SearchContests :many
SELECT id, title, start_at, duration_seconds, rate_change
FROM contests
WHERE lower(id) LIKE sqlc.arg(pattern)
   OR regexp_replace(lower(title), '\s', '', 'g') LIKE sqlc.arg(pattern)
ORDER BY start_at DESC, id
LIMIT 30;

-- name: FindContestByID :one
SELECT id, title, start_at, duration_seconds, rate_change
FROM contests
WHERE id = $1;

-- name: ListContestProblems :many
SELECT problem.id, contest_problem.problem_index, problem.name
FROM contest_problems AS contest_problem
JOIN problems AS problem ON problem.id = contest_problem.problem_id
WHERE contest_problem.contest_id = $1
ORDER BY
    char_length(contest_problem.problem_index),
    contest_problem.problem_index,
    problem.id;

-- name: FindProblemsByIDs :many
SELECT id, contest_id, problem_index, name
FROM problems
WHERE id = ANY(sqlc.arg(ids)::text[]);

-- name: FindProblemByID :one
SELECT id, contest_id, problem_index, name
FROM problems
WHERE id = $1;

-- name: InsertProblemIfNotExists :exec
INSERT INTO problems (id, contest_id)
VALUES ($1, $2)
ON CONFLICT (id) DO NOTHING;

-- name: UpsertContests :exec
INSERT INTO contests (id, title, start_at, duration_seconds, rate_change)
SELECT
    unnest(sqlc.arg(ids)::text[]),
    unnest(sqlc.arg(titles)::text[]),
    unnest(sqlc.arg(start_ats)::timestamptz[]),
    unnest(sqlc.arg(durations_seconds)::bigint[]),
    unnest(sqlc.arg(rate_changes)::text[])
ON CONFLICT (id) DO UPDATE
SET
    title = EXCLUDED.title,
    start_at = EXCLUDED.start_at,
    duration_seconds = EXCLUDED.duration_seconds,
    rate_change = EXCLUDED.rate_change,
    updated_at = now()
WHERE (contests.title, contests.start_at, contests.duration_seconds, contests.rate_change)
    IS DISTINCT FROM (EXCLUDED.title, EXCLUDED.start_at, EXCLUDED.duration_seconds, EXCLUDED.rate_change);

-- name: UpsertProblems :exec
INSERT INTO problems (id, contest_id, problem_index, name)
SELECT
    unnest(sqlc.arg(ids)::text[]),
    unnest(sqlc.arg(contest_ids)::text[]),
    unnest(sqlc.arg(problem_indexes)::text[]),
    unnest(sqlc.arg(names)::text[])
ON CONFLICT (id) DO UPDATE
SET
    contest_id = EXCLUDED.contest_id,
    problem_index = EXCLUDED.problem_index,
    name = EXCLUDED.name,
    updated_at = now()
WHERE (problems.contest_id, problems.problem_index, problems.name)
    IS DISTINCT FROM (EXCLUDED.contest_id, EXCLUDED.problem_index, EXCLUDED.name);

-- name: UpsertContestProblems :exec
INSERT INTO contest_problems (contest_id, problem_id, problem_index)
SELECT
    unnest(sqlc.arg(contest_ids)::text[]),
    unnest(sqlc.arg(problem_ids)::text[]),
    unnest(sqlc.arg(problem_indexes)::text[])
ON CONFLICT (contest_id, problem_id) DO UPDATE
SET problem_index = EXCLUDED.problem_index
WHERE contest_problems.problem_index IS DISTINCT FROM EXCLUDED.problem_index;
