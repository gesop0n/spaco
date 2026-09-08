import { Link } from "@tanstack/react-router";
import { BookOpen, CalendarDays, History, Pause, Play, Plus, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { LearningPageLayout } from "../LearningPageLayout";
import { ProblemIdentity } from "../ProblemIdentity";
import { dueLabel, formatDay, resultLabels } from "../ReviewWorkspace/model";
import { useReviewListPage } from "./_.hook";

const filters = [
  { value: "all", label: "すべて" },
  { value: "active", label: "復習中" },
  { value: "paused", label: "一時停止" },
] as const;

export function ReviewListPage() {
  const page = useReviewListPage();
  return (
    <LearningPageLayout
      eyebrow="YOUR PRACTICE LIBRARY"
      title="復習リスト"
      description="積み重ねた記録と、これからの予定。自分のペースで続けよう。"
      actions={
        <Link
          to="/problems"
          className="inline-flex min-h-11 items-center gap-2 rounded-lg bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground shadow-action hover:bg-accent-hover"
        >
          <Plus className="size-4" aria-hidden="true" />
          問題を追加
        </Link>
      }
    >
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div
          className="flex gap-1 rounded-xl border border-border bg-card p-1"
          aria-label="復習状態で絞り込み"
        >
          {filters.map(({ value, label }) => (
            <button
              key={value}
              type="button"
              aria-pressed={page.filter === value}
              onClick={() => page.setFilter(value)}
              className={`min-h-10 rounded-lg px-3 text-xs font-medium transition-colors sm:px-4 sm:text-sm ${page.filter === value ? "bg-accent text-primary" : "text-muted-foreground hover:bg-muted"}`}
            >
              {label}
              <span className="ml-1.5 text-[11px]">{page.counts[value]}</span>
            </button>
          ))}
        </div>
        <div className="relative w-full sm:w-72">
          <Search
            className="absolute top-3 left-3 size-4 text-muted-foreground"
            aria-hidden="true"
          />
          <input
            aria-label="登録した問題を検索"
            type="search"
            value={page.query}
            onChange={(event) => page.setQuery(event.target.value)}
            placeholder="問題名・IDで検索"
            className="h-10 w-full rounded-lg border border-input bg-card pr-3 pl-9 text-sm"
          />
        </div>
      </div>
      <p className="-mt-2 text-xs leading-6 text-muted-foreground">
        一時停止しても履歴は残ります。再開すると、予定日を過ぎた問題は今日の復習に戻ります。
      </p>
      <div className="grid gap-4">
        {page.visible.map((problem) => (
          <article
            key={problem.id}
            aria-label={problem.title}
            className="min-w-0 overflow-hidden rounded-2xl border border-border bg-card shadow-card"
          >
            <div className="p-5 sm:p-6">
              <div className="flex flex-wrap items-center justify-between gap-4">
                <ProblemIdentity problem={problem} />
                <span
                  className={`inline-flex items-center gap-1.5 rounded-md px-2.5 py-1 text-[11px] ${problem.paused ? "bg-muted text-muted-foreground" : problem.dueOn < page.today ? "bg-amber-50 text-amber-800 dark:bg-amber-950 dark:text-amber-200" : "bg-accent text-primary"}`}
                >
                  <CalendarDays className="size-3.5" aria-hidden="true" />
                  {problem.paused
                    ? "一時停止中"
                    : problem.dueOn < page.today
                      ? `${formatDay(problem.dueOn)}予定 · 期限超過`
                      : `次の復習：${dueLabel(problem.dueOn, page.today)}`}
                </span>
              </div>
              <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
                <p className="text-xs leading-6 text-muted-foreground">
                  復習 {problem.history.length}回<span className="mx-2">·</span>
                  {problem.history.length
                    ? resultLabels[problem.history.at(-1)!.result]
                    : "まだ復習していません"}
                </p>
                <div className="flex flex-wrap items-center gap-2">
                  <Button
                    variant="ghost"
                    className="min-h-10 gap-1.5 text-xs text-muted-foreground"
                    onClick={() => page.togglePause(problem.id)}
                  >
                    {problem.paused ? <Play aria-hidden="true" /> : <Pause aria-hidden="true" />}
                    {problem.paused ? "再開" : "一時停止"}
                  </Button>
                  {!problem.paused && (
                    <Link
                      to="/reviews"
                      search={{ problem: problem.id }}
                      className="inline-flex min-h-10 items-center rounded-lg border border-border px-4 text-xs font-semibold text-primary hover:bg-accent"
                    >
                      取り組む
                    </Link>
                  )}
                </div>
              </div>
            </div>
            <details className="group border-t border-border bg-muted/25">
              <summary className="cursor-pointer px-5 py-3 text-xs font-medium text-muted-foreground hover:text-foreground sm:px-6">
                <span className="ml-1 inline-flex items-center gap-2">
                  <History className="size-3.5" aria-hidden="true" />
                  登録メモ・復習の履歴
                </span>
              </summary>
              <div className="grid gap-4 px-5 pt-2 pb-5 sm:px-6">
                <div className="rounded-xl border border-border bg-card p-4">
                  <p className="text-[11px] text-muted-foreground">
                    {formatDay(problem.registeredOn)}に登録
                  </p>
                  <p className="mt-2 whitespace-pre-wrap break-words text-xs leading-6">
                    {problem.registrationNote || "登録メモはありません。"}
                  </p>
                </div>
                {problem.history.length ? (
                  <ol className="grid gap-3" aria-label="復習の履歴">
                    {problem.history.toReversed().map((entry) => (
                      <li key={entry.id} className="border-l-2 border-primary/25 py-1 pl-4">
                        <p className="text-[11px] text-muted-foreground">
                          {formatDay(entry.performedAt.slice(0, 10))}{" "}
                          {entry.performedAt.slice(11, 16)} · {page.timeZone}
                        </p>
                        <p className="mt-1 text-xs font-semibold">{resultLabels[entry.result]}</p>
                        {entry.note && (
                          <p className="mt-2 whitespace-pre-wrap break-words text-xs leading-6 text-muted-foreground">
                            {entry.note}
                          </p>
                        )}
                      </li>
                    ))}
                  </ol>
                ) : (
                  <p className="px-1 text-xs text-muted-foreground">
                    まだ復習の記録はありません。登録は復習回数に含まれません。
                  </p>
                )}
              </div>
            </details>
          </article>
        ))}
        {!page.visible.length && (
          <div className="rounded-2xl border border-dashed border-border bg-card/70 px-5 py-14 text-center">
            <BookOpen className="mx-auto size-7 text-primary" aria-hidden="true" />
            <h2 className="mt-4 font-semibold">
              {page.query || page.filter !== "all"
                ? "条件に合う問題はありません"
                : "復習したい問題を集めよう"}
            </h2>
            <p className="mt-2 text-sm text-muted-foreground">
              {page.query || page.filter !== "all"
                ? "検索語や絞り込みを変えてみてください。"
                : "コンテストや問題URLから登録できます。"}
            </p>
            <Link
              to="/problems"
              className="mt-5 inline-flex min-h-10 items-center text-sm font-semibold text-primary"
            >
              問題を登録する →
            </Link>
          </div>
        )}
      </div>
    </LearningPageLayout>
  );
}
