import assert from "node:assert/strict";
import { test } from "node:test";
import {
  addDays,
  dueProblems,
  isValidLocalDateTime,
  localDateTime,
  problemFromUrl,
  problemUrl,
  studyReducer,
} from "../model.ts";
import type { Problem, ReviewResult } from "../types.ts";

const today = "2026-09-08";
const problem: Problem = { id: "abc350_a", contestId: "abc350", index: "A", title: "Past ABCs" };
const register = (problems = [problem]) =>
  studyReducer([], { type: "register", problems, today, note: "  登録メモ  " });

test("暦日の加算は月末・年末・うるう年に対応する", () => {
  assert.equal(addDays("2026-12-31", 1), "2027-01-01");
  assert.equal(addDays("2024-02-28", 1), "2024-02-29");
  assert.equal(addDays("2026-02-28", 1), "2026-03-01");
});

test("実施日時はアカウントのタイムゾーンを使う", () => {
  const now = new Date("2026-09-08T15:30:00Z");
  assert.equal(localDateTime("Asia/Tokyo", now), "2026-09-09T00:30");
  assert.equal(localDateTime("America/Los_Angeles", now), "2026-09-08T08:30");
});

test("存在しない日付・時刻や不正な形式を拒否する", () => {
  assert(isValidLocalDateTime("2024-02-29T23:59"));
  for (const value of [
    "2026-02-29T12:00",
    "2026-09-08T24:00",
    "2026-09-08",
    "",
    "2026-09-08T12:00Z",
  ]) {
    assert.equal(isValidLocalDateTime(value), false, value);
  }
});

test("AtCoderのHTTPS問題URLだけ受け入れ、リンクを正規化する", () => {
  const parsed = problemFromUrl(
    " https://atcoder.jp/contests/abc350/tasks/abc350_a/?lang=ja#task-statement ",
    [problem],
  );
  assert.deepEqual(parsed, problem);
  assert.equal(problemUrl(parsed!), "https://atcoder.jp/contests/abc350/tasks/abc350_a");
  assert.equal(
    problemFromUrl("https://atcoder.jp/contests/abc001/tasks/abc001_1", [])?.title,
    "abc001_1",
  );
  for (const url of [
    "javascript:alert(1)",
    "http://atcoder.jp/contests/abc350/tasks/abc350_a",
    "https://atcoder.jp.evil.test/contests/abc350/tasks/abc350_a",
    "https://atcoder.jp@evil.test/contests/abc350/tasks/abc350_a",
    "https://user:pass@atcoder.jp/contests/abc350/tasks/abc350_a",
    "https://atcoder.jp/contests/abc350",
    "https://atcoder.jp/contests/abc350/tasks/a/extra",
  ]) {
    assert.equal(problemFromUrl(url, []), undefined, url);
  }
});

test("登録は翌日の予定を作り、復習履歴を増やさず、重複を除外する", () => {
  const state = register([problem, problem]);
  assert.equal(state.length, 1);
  assert.equal(state[0].dueOn, "2026-09-09");
  assert.equal(state[0].registrationNote, "登録メモ");
  assert.deepEqual(state[0].history, []);
  assert.deepEqual(
    studyReducer(state, { type: "register", problems: [problem], today, note: "上書きしない" }),
    state,
  );
});

for (const [result, expected] of Object.entries({
  independent: "2026-09-17",
  assisted: "2026-09-13",
  retry: "2026-09-11",
})) {
  test(`${result}の仮の次回予定は元の予定日ではなく実施日から計算する`, () => {
    const entry = {
      id: "entry",
      result: result as ReviewResult,
      performedAt: "2026-09-10T20:00",
      note: "結果メモ",
    };
    const state = studyReducer(register(), { type: "record", problemId: problem.id, entry });
    assert.equal(state[0].dueOn, expected);
    assert.deepEqual(state[0].history, [entry]);
    assert.equal(state[0].registrationNote, "登録メモ");
    assert.deepEqual(studyReducer(state, { type: "record", problemId: problem.id, entry }), state);
  });
}

test("期限超過は古い順に残り、一時停止・再開でも履歴や予定を変えない", () => {
  const state = [
    { ...register()[0], dueOn: today },
    { ...register()[0], id: "older", dueOn: "2026-09-06" },
    { ...register()[0], id: "future", dueOn: "2026-09-10" },
  ];
  assert.deepEqual(
    dueProblems(state, today).map(({ id }) => id),
    ["older", problem.id],
  );
  const paused = studyReducer(state, { type: "togglePause", problemId: "older" });
  assert.deepEqual(
    dueProblems(paused, today).map(({ id }) => id),
    [problem.id],
  );
  const entry = { id: "entry", result: "retry" as const, performedAt: `${today}T12:00`, note: "" };
  assert.deepEqual(studyReducer(paused, { type: "record", problemId: "older", entry }), paused);
  assert.deepEqual(studyReducer(paused, { type: "togglePause", problemId: "older" }), state);
  assert.deepEqual(studyReducer(state, { type: "record", problemId: "missing", entry }), state);
});
