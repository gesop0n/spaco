import { Link } from "@tanstack/react-router";
import {
  ArrowRight,
  CalendarDays,
  CheckCircle2,
  CirclePlay,
  Clock3,
  Plus,
  RotateCcw,
  Sparkles,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { LearningPageLayout } from "../LearningPageLayout";
import { ProblemIdentity } from "../ProblemIdentity";
import { ReviewResultSheet } from "../ReviewResultSheet";
import { itemProblem, logLabel, performedOn } from "../ReviewWorkspace/api";
import { dueLabel, formatDay, problemTitle } from "../ReviewWorkspace/model";
import { useReviewDashboardPage } from "./_.hook";

export function ReviewDashboardPage() {
  const page = useReviewDashboardPage();
  return (
    <LearningPageLayout
      eyebrow="A LITTLE PRACTICE, EVERY DAY"
      title="今日の復習"
      description={
        page.today
          ? `${formatDay(page.today)} · ${page.timeZone} — ひとつずつ、「解ける」を増やそう。`
          : "ひとつずつ、「解ける」を増やそう。"
      }
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
      {page.error ? (
        <div
          role="alert"
          className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-6 text-sm"
        >
          <p>{page.error}</p>
          <Button variant="secondary" className="h-10 gap-2 px-4" onClick={page.retry}>
            <RotateCcw aria-hidden="true" />
            再読み込み
          </Button>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-3 gap-3 sm:gap-4">
            {[
              { label: "今日の対象", value: page.due.length, icon: CalendarDays },
              { label: "期限超過", value: page.overdue, icon: Clock3 },
              { label: "今日の記録", value: page.completed, icon: CheckCircle2 },
            ].map(({ label, value, icon: Icon }) => (
              <div key={label} className="rounded-2xl border border-border bg-card/90 p-4 sm:p-5">
                <p className="flex items-center gap-2 text-[11px] text-muted-foreground sm:text-xs">
                  <Icon className="hidden size-4 text-primary sm:block" aria-hidden="true" />
                  {label}
                </p>
                {page.loading ? (
                  <Skeleton className="mt-3 h-9 w-14" />
                ) : (
                  <p className="mt-3 text-3xl font-semibold tracking-tight">
                    {value}
                    <span className="ml-1 text-xs font-normal text-muted-foreground">問</span>
                  </p>
                )}
              </div>
            ))}
          </div>
          {page.invalidSelection && (
            <div role="alert" className="rounded-xl border border-border bg-card p-4 text-sm">
              この問題は未登録、または一時停止中です。
              <Link to="/review-list" className="ml-2 text-primary underline underline-offset-4">
                復習リストを確認
              </Link>
              <Button variant="ghost" className="ml-2" onClick={page.close}>
                閉じる
              </Button>
            </div>
          )}
          {page.selectionError && (
            <div role="alert" className="rounded-xl border border-border bg-card p-4 text-sm">
              {page.selectionError}
              <Button variant="ghost" className="ml-2" onClick={page.close}>
                閉じる
              </Button>
            </div>
          )}
          <div className="grid items-start gap-6 xl:grid-cols-[minmax(0,1fr)_272px]">
            <section
              className="min-w-0 overflow-hidden rounded-2xl border border-border bg-card shadow-card"
              aria-labelledby="review-queue-title"
            >
              <header className="border-b border-border p-5 sm:p-6">
                <div className="flex items-center justify-between gap-3">
                  <h2 id="review-queue-title" className="font-semibold">
                    今日のキュー
                  </h2>
                  <span className="text-xs text-muted-foreground">予定日が古い順</span>
                </div>
                <div
                  className="mt-4 h-1.5 overflow-hidden rounded-full bg-muted"
                  role="progressbar"
                  aria-label="今日の復習の進み具合"
                  aria-valuemin={0}
                  aria-valuemax={100}
                  aria-valuenow={page.progress}
                >
                  <div
                    className="h-full rounded-full bg-primary transition-[width]"
                    style={{ width: `${page.progress}%` }}
                  />
                </div>
                <p className="mt-2 text-[11px] text-muted-foreground">
                  {page.completed
                    ? `${page.completed}件を記録しました。あと${page.due.length}問。`
                    : "まだ取り組んでいない問題は、期限を過ぎてもここに残ります。"}
                </p>
              </header>
              <div className="divide-y divide-border">
                {page.loading ? (
                  <div className="grid gap-4 p-5 sm:p-6" aria-label="今日の復習を読み込み中">
                    {[0, 1, 2].map((index) => (
                      <Skeleton key={index} className="h-20" />
                    ))}
                  </div>
                ) : (
                  <>
                    {page.due.map((item) => {
                      const problem = itemProblem(item);
                      const overdue = item.dueOn < page.today;
                      return (
                        <article key={item.id} className="p-5 sm:p-6">
                          <div className="mb-3 flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                            <span
                              className={`rounded-md px-2 py-0.5 ${overdue ? "bg-amber-50 text-amber-800 dark:bg-amber-950 dark:text-amber-200" : "bg-accent text-primary"}`}
                            >
                              {overdue ? `${formatDay(item.dueOn)}予定 · 期限超過` : "今日の予定"}
                            </span>
                            <span>復習 {item.reviewCount}回</span>
                          </div>
                          <div className="flex flex-wrap items-center justify-between gap-4">
                            <ProblemIdentity problem={problem} />
                            <Button
                              onClick={() => page.start(problem.id)}
                              className="ml-auto h-10 gap-2 px-4"
                              aria-label={`${problemTitle(problem)}に取り組む`}
                            >
                              <CirclePlay aria-hidden="true" />
                              取り組む
                            </Button>
                          </div>
                          <p className="mt-4 text-xs text-muted-foreground">
                            前回：
                            {item.lastLog
                              ? `${logLabel(item.lastLog)} · ${formatDay(performedOn(item.lastLog, page.timeZone))}`
                              : "初めての復習"}
                          </p>
                          {item.registrationNote && (
                            <p className="mt-2 break-words text-xs leading-6 text-muted-foreground">
                              {item.registrationNote}
                            </p>
                          )}
                        </article>
                      );
                    })}
                    {!page.due.length && (
                      <div className="flex flex-col items-center px-6 py-14 text-center">
                        <span className="grid size-14 place-items-center rounded-full bg-accent text-primary">
                          <CheckCircle2 className="size-6" aria-hidden="true" />
                        </span>
                        <h3 className="mt-5 text-lg font-semibold">
                          {page.completed ? "今日の復習、おつかれさま。" : "今日の復習はありません"}
                        </h3>
                        <p className="mt-3 text-sm leading-6 text-muted-foreground">
                          {page.upcoming.length
                            ? "次の予定まで、ひと息。復習リストから予定を確認できます。"
                            : "気になる問題を登録して、次の復習につなげましょう。"}
                        </p>
                        <Link to="/problems" className="mt-6 text-sm font-semibold text-primary">
                          問題を登録する →
                        </Link>
                      </div>
                    )}
                  </>
                )}
              </div>
            </section>
            <aside className="grid gap-5">
              <section className="rounded-2xl border border-primary/15 bg-aurora p-5">
                <Sparkles className="size-5 text-primary" aria-hidden="true" />
                <h2 className="mt-3 text-sm font-semibold">思い出すことから、はじめよう。</h2>
                <p className="mt-3 text-xs leading-7 text-muted-foreground">
                  まずは解説を見ずに、考え方を組み立てる。
                  <br />
                  難しければ、ヒントを見ても大丈夫。
                  <br />
                  今回の結果を、次の一歩にしましょう。
                </p>
              </section>
              <section className="rounded-2xl border border-border bg-card p-5">
                <h2 className="text-sm font-semibold">この先の復習</h2>
                <div className="mt-4 grid gap-4">
                  {page.upcoming.map((item) => {
                    const problem = itemProblem(item);
                    return (
                      <div key={item.id}>
                        <p className="text-[11px] font-medium text-primary">
                          {dueLabel(item.dueOn, page.today)}
                        </p>
                        <p className="mt-1 break-words text-sm font-medium">
                          {problemTitle(problem)}
                        </p>
                        <p className="mt-1 text-[11px] text-muted-foreground">
                          {problem.contestId.toUpperCase()} · {problem.problemIndex ?? problem.id}
                        </p>
                      </div>
                    );
                  })}
                  {!page.loading && !page.upcoming.length && (
                    <p className="text-xs leading-6 text-muted-foreground">
                      結果を記録すると、ここに次の予定が表示されます。
                    </p>
                  )}
                </div>
                <Link
                  to="/review-list"
                  className="mt-5 inline-flex items-center gap-2 text-xs font-semibold text-primary"
                >
                  すべての復習を見る
                  <ArrowRight className="size-3.5" aria-hidden="true" />
                </Link>
              </section>
            </aside>
          </div>
        </>
      )}
      {page.selected && page.timeZone && (
        <ReviewResultSheet
          key={page.selected.id}
          item={page.selected}
          timeZone={page.timeZone}
          onClose={page.close}
        />
      )}
    </LearningPageLayout>
  );
}
