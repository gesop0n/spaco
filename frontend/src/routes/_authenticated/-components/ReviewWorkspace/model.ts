import type { DifficultyKey, ProblemReference, ProblemSummary, ResultKey } from "./types";

export const resultLabels: Record<ResultKey, string> = {
  independent: "自力でACできた",
  assisted: "解説・ヒントを見てACした",
  unsolved: "ACできなかった",
};

export const difficultyLabels: Record<DifficultyKey, string> = {
  hard: "苦戦した",
  good: "普通",
  easy: "余裕だった",
};

/** 結果と、自力でACできたときの手応えを1つの表示にまとめる。 */
export function outcomeLabel(result: ResultKey, difficulty?: DifficultyKey): string {
  if (result !== "independent" || !difficulty) return resultLabels[result];
  return `${resultLabels[result]}（${difficultyLabels[difficulty]}）`;
}

export function problemTitle(problem: ProblemSummary): string {
  return problem.name || problem.id;
}

export function addDays(day: string, days: number): string {
  const date = new Date(`${day}T00:00:00Z`);
  date.setUTCDate(date.getUTCDate() + days);
  return date.toISOString().slice(0, 10);
}

/** YYYY-MM-DD形式の2つの暦日の差を日数で返す。 */
export function daysBetween(from: string, to: string): number {
  return Math.round((Date.parse(`${to}T00:00:00Z`) - Date.parse(`${from}T00:00:00Z`)) / 86_400_000);
}

function zonedParts(timeZone: string, date: Date) {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hourCycle: "h23",
  }).formatToParts(date);
  const value = (name: string) => parts.find((part) => part.type === name)?.value ?? "";
  return {
    year: value("year"),
    month: value("month"),
    day: value("day"),
    hour: value("hour"),
    minute: value("minute"),
    second: value("second"),
  };
}

/** timeZoneでの日時をYYYY-MM-DDTHH:mm形式で返す。datetime-localの値に使う。 */
export function localDateTime(timeZone: string, now = new Date()): string {
  const { year, month, day, hour, minute } = zonedParts(timeZone, now);
  return `${year}-${month}-${day}T${hour}:${minute}`;
}

/** timeZoneで、dateが属する暦日を返す。 */
export function localDateOf(date: Date, timeZone: string): string {
  return localDateTime(timeZone, date).slice(0, 10);
}

/** その時点で、timeZoneの時刻がUTCから何ミリ秒進んでいるかを返す。 */
function timeZoneOffset(instant: number, timeZone: string): number {
  const parts = zonedParts(timeZone, new Date(instant));
  const wallClock = Date.UTC(
    Number(parts.year),
    Number(parts.month) - 1,
    Number(parts.day),
    Number(parts.hour),
    Number(parts.minute),
    Number(parts.second),
  );
  return wallClock - Math.floor(instant / 1000) * 1000;
}

/**
 * timeZoneでのYYYY-MM-DDTHH:mm形式の日時を、絶対時刻に変換する。
 * 夏時間の切り替え前後でも正しい時差を使うため、変換後の時刻でもう一度時差を確認する。
 */
export function zonedDateTimeToDate(value: string, timeZone: string): Date {
  const wallClock = Date.parse(`${value}:00Z`);
  const firstGuess = wallClock - timeZoneOffset(wallClock, timeZone);
  const offset = timeZoneOffset(firstGuess, timeZone);
  return new Date(wallClock - offset);
}

export function formatDay(day: string): string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(day)) return day;
  return new Intl.DateTimeFormat("ja-JP", {
    timeZone: "UTC",
    month: "long",
    day: "numeric",
    weekday: "short",
  }).format(new Date(`${day}T00:00:00Z`));
}

export function isValidLocalDateTime(value: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(value)) return false;
  const date = new Date(`${value}Z`);
  return !Number.isNaN(date.getTime()) && date.toISOString().slice(0, 16) === value;
}

export function dueLabel(day: string, today: string): string {
  if (day === today) return "今日";
  if (day === addDays(today, 1)) return "明日";
  return formatDay(day);
}

/**
 * AtCoderの問題URLの形だけを確認する。問題の実在確認と正規化はサーバーが行う。
 * APG4bのように大文字を含むIDも受け入れる。
 */
export function parseProblemUrl(value: string): ProblemReference | undefined {
  try {
    const url = new URL(value.trim());
    if (url.origin !== "https://atcoder.jp" || url.username || url.password) return;
    const match = url.pathname.match(/^\/contests\/([A-Za-z0-9_-]+)\/tasks\/([A-Za-z0-9_-]+)\/?$/);
    if (!match) return;
    const [, contestId, problemId] = match;
    return { contestId, problemId };
  } catch {
    return;
  }
}
