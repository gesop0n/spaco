import type { Problem, ReviewResult, StudyAction, StudyProblem } from "./types";

export const resultLabels: Record<ReviewResult, string> = {
  independent: "自力でACできた",
  assisted: "解説・ヒントを見てACした",
  retry: "ACできなかった",
};

/** UI確認用の仮の間隔。実際のスケジューリングには使用しない。 */
export const previewIntervals: Record<ReviewResult, number> = {
  independent: 7,
  assisted: 3,
  retry: 1,
};

export function addDays(day: string, days: number): string {
  const date = new Date(`${day}T00:00:00Z`);
  date.setUTCDate(date.getUTCDate() + days);
  return date.toISOString().slice(0, 10);
}

export function localDateTime(timeZone: string, now = new Date()): string {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }).formatToParts(now);
  const value = (name: string) => parts.find((part) => part.type === name)?.value;
  return `${value("year")}-${value("month")}-${value("day")}T${value("hour")}:${value("minute")}`;
}

export function formatDay(day: string): string {
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

export function problemUrl(problem: Problem): string {
  return `https://atcoder.jp/contests/${problem.contestId}/tasks/${problem.id}`;
}

/** URLの形だけを検証する。存在確認やAtCoderへのアクセスは行わない。 */
export function problemFromUrl(value: string, catalog: Problem[]): Problem | undefined {
  try {
    const url = new URL(value.trim());
    if (url.origin !== "https://atcoder.jp" || url.username || url.password) return;
    const match = url.pathname.match(/^\/contests\/([a-z0-9_-]+)\/tasks\/([a-z0-9_-]+)\/?$/);
    if (!match) return;
    const [, contestId, id] = match;
    return (
      catalog.find((problem) => problem.id === id) ?? {
        id,
        contestId,
        index: id.split("_").at(-1)?.slice(0, 2).toUpperCase() || "?",
        title: id,
      }
    );
  } catch {
    return;
  }
}

export function studyReducer(state: StudyProblem[], action: StudyAction): StudyProblem[] {
  switch (action.type) {
    case "register": {
      const registered = new Set(state.map((problem) => problem.id));
      const additions = action.problems
        .filter((problem) => {
          if (registered.has(problem.id)) return false;
          registered.add(problem.id);
          return true;
        })
        .map((problem) => ({
          ...problem,
          registeredOn: action.today,
          dueOn: addDays(action.today, 1),
          paused: false,
          registrationNote: action.note.trim(),
          history: [],
        }));
      return [...state, ...additions];
    }
    case "record":
      return state.map((problem) => {
        if (
          problem.id !== action.problemId ||
          problem.paused ||
          problem.history.some(({ id }) => id === action.entry.id)
        )
          return problem;
        return {
          ...problem,
          dueOn: addDays(
            action.entry.performedAt.slice(0, 10),
            previewIntervals[action.entry.result],
          ),
          history: [...problem.history, action.entry],
        };
      });
    case "togglePause":
      return state.map((problem) =>
        problem.id === action.problemId ? { ...problem, paused: !problem.paused } : problem,
      );
  }
}

export function dueProblems(problems: StudyProblem[], today: string): StudyProblem[] {
  return problems
    .filter((problem) => !problem.paused && problem.dueOn <= today)
    .sort((a, b) => a.dueOn.localeCompare(b.dueOn) || a.id.localeCompare(b.id));
}
