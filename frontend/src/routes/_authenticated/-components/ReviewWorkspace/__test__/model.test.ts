import assert from "node:assert/strict";
import { test } from "node:test";
import {
  addDays,
  daysBetween,
  dueLabel,
  isValidLocalDateTime,
  localDateOf,
  localDateTime,
  outcomeLabel,
  parseProblemUrl,
  problemTitle,
  zonedDateTimeToDate,
} from "../model.ts";

test("暦日の加算と差は月末・年末・うるう年に対応する", () => {
  assert.equal(addDays("2026-12-31", 1), "2027-01-01");
  assert.equal(addDays("2024-02-28", 1), "2024-02-29");
  assert.equal(addDays("2026-02-28", 1), "2026-03-01");
  assert.equal(daysBetween("2026-09-08", "2026-09-11"), 3);
  assert.equal(daysBetween("2024-02-28", "2024-03-01"), 2);
  assert.equal(daysBetween("2026-09-11", "2026-09-11"), 0);
});

test("実施日時はアカウントのタイムゾーンで表示する", () => {
  const now = new Date("2026-09-08T15:30:00Z");
  assert.equal(localDateTime("Asia/Tokyo", now), "2026-09-09T00:30");
  assert.equal(localDateTime("America/Los_Angeles", now), "2026-09-08T08:30");
  assert.equal(localDateOf(now, "Asia/Tokyo"), "2026-09-09");
  assert.equal(localDateOf(now, "America/Los_Angeles"), "2026-09-08");
});

test("タイムゾーンでの日時を絶対時刻に変換する", () => {
  assert.equal(
    zonedDateTimeToDate("2026-09-09T00:30", "Asia/Tokyo").toISOString(),
    "2026-09-08T15:30:00.000Z",
  );
  assert.equal(
    zonedDateTimeToDate("2026-09-08T08:30", "America/Los_Angeles").toISOString(),
    "2026-09-08T15:30:00.000Z",
  );
  // 2026-03-08 02:00に夏時間へ切り替わった直後は、UTC-4で変換する。
  assert.equal(
    zonedDateTimeToDate("2026-03-08T03:30", "America/New_York").toISOString(),
    "2026-03-08T07:30:00.000Z",
  );
  assert.equal(
    zonedDateTimeToDate("2026-03-07T23:30", "America/New_York").toISOString(),
    "2026-03-08T04:30:00.000Z",
  );
  for (const timeZone of ["Asia/Tokyo", "Europe/London", "America/Los_Angeles"]) {
    const value = "2026-07-01T12:05";
    assert.equal(localDateTime(timeZone, zonedDateTimeToDate(value, timeZone)), value, timeZone);
  }
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

test("予定日は今日・明日を言葉で表示する", () => {
  assert.equal(dueLabel("2026-09-08", "2026-09-08"), "今日");
  assert.equal(dueLabel("2026-09-09", "2026-09-08"), "明日");
  assert.notEqual(dueLabel("2026-09-10", "2026-09-08"), "明日");
});

test("AtCoderのHTTPS問題URLだけ受け入れ、大文字を含むIDも読み取る", () => {
  assert.deepEqual(
    parseProblemUrl(" https://atcoder.jp/contests/abc350/tasks/abc350_a/?lang=ja#task-statement "),
    { contestId: "abc350", problemId: "abc350_a" },
  );
  assert.deepEqual(parseProblemUrl("https://atcoder.jp/contests/APG4b/tasks/APG4b_a"), {
    contestId: "APG4b",
    problemId: "APG4b_a",
  });
  for (const url of [
    "javascript:alert(1)",
    "http://atcoder.jp/contests/abc350/tasks/abc350_a",
    "https://atcoder.jp.evil.test/contests/abc350/tasks/abc350_a",
    "https://atcoder.jp@evil.test/contests/abc350/tasks/abc350_a",
    "https://user:pass@atcoder.jp/contests/abc350/tasks/abc350_a",
    "https://atcoder.jp/contests/abc350",
    "https://atcoder.jp/contests/abc350/tasks/a/extra",
  ]) {
    assert.equal(parseProblemUrl(url), undefined, url);
  }
});

test("結果の表示には、自力ACのときだけ手応えを添える", () => {
  assert.equal(outcomeLabel("independent", "hard"), "自力でACできた（苦戦した）");
  assert.equal(outcomeLabel("independent"), "自力でACできた");
  assert.equal(outcomeLabel("assisted", "easy"), "解説・ヒントを見てACした");
  assert.equal(outcomeLabel("unsolved"), "ACできなかった");
});

test("問題名が未補完なら問題IDを表示する", () => {
  assert.equal(
    problemTitle({ id: "abc350_d", contestId: "abc350", name: "New Friends" }),
    "New Friends",
  );
  assert.equal(problemTitle({ id: "abc999_a", contestId: "abc999" }), "abc999_a");
});
