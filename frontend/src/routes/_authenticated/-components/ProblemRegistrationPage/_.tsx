import { Link } from "@tanstack/react-router";
import {
  ArrowRight,
  CalendarDays,
  Check,
  CheckCircle2,
  ExternalLink,
  Link2,
  Plus,
  Search,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { LearningPageLayout } from "../LearningPageLayout";
import { ProblemIdentity } from "../ProblemIdentity";
import { formatDay, problemUrl } from "../ReviewWorkspace/model";
import { useProblemRegistrationPage } from "./_.hook";

const inputClassName =
  "min-h-11 w-full rounded-lg border border-input bg-card px-3 py-2.5 text-sm outline-none focus:border-primary focus:ring-3 focus:ring-primary/15 aria-invalid:border-danger";

export function ProblemRegistrationPage() {
  const page = useProblemRegistrationPage();
  return (
    <LearningPageLayout
      eyebrow="ADD TO YOUR PRACTICE"
      title="問題を登録"
      description="解説を読んだ問題も、もう一度挑戦したい問題も。次の復習につなげましょう。"
    >
      {page.notice && (
        <div
          role="status"
          className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-primary/20 bg-accent/70 p-5"
        >
          <div className="flex items-start gap-3">
            <CheckCircle2 className="mt-0.5 size-5 shrink-0 text-primary" aria-hidden="true" />
            <div>
              <p className="text-sm font-semibold">
                {page.notice.count}件を復習リストに追加しました
              </p>
              <p className="mt-1 text-xs text-muted-foreground">
                最初の復習は {formatDay(page.nextDay)}。登録は復習回数に含めません。
              </p>
            </div>
          </div>
          <div className="flex flex-wrap gap-3 text-xs font-semibold text-primary">
            <Link to="/review-list" className="rounded-md px-2 py-2 hover:bg-card">
              復習リストを見る →
            </Link>
            {page.notice.count === 1 && (
              <Link
                to="/reviews"
                search={{ problem: page.notice.firstId }}
                className="rounded-md px-2 py-2 hover:bg-card"
              >
                今すぐ取り組む →
              </Link>
            )}
          </div>
        </div>
      )}
      <div
        className="flex w-fit flex-wrap gap-1 rounded-xl border border-border bg-card/80 p-1"
        aria-label="登録方法"
      >
        <Button
          variant={page.mode === "contest" ? "secondary" : "ghost"}
          className="h-10 gap-2 px-4"
          aria-pressed={page.mode === "contest"}
          onClick={() => page.setMode("contest")}
        >
          <Search aria-hidden="true" />
          コンテストから探す
        </Button>
        <Button
          variant={page.mode === "url" ? "secondary" : "ghost"}
          className="h-10 gap-2 px-4"
          aria-pressed={page.mode === "url"}
          onClick={() => page.setMode("url")}
        >
          <Link2 aria-hidden="true" />
          URLから登録
        </Button>
      </div>

      {page.mode === "contest" ? (
        <div className="grid items-start gap-5 xl:grid-cols-[240px_minmax(0,1fr)]">
          <aside className="rounded-2xl border border-border bg-card p-4">
            <label htmlFor="contest-search" className="mb-3 block text-xs font-semibold">
              コンテストを探す
            </label>
            <div className="relative">
              <Search
                className="pointer-events-none absolute top-3.5 left-3 size-4 text-muted-foreground"
                aria-hidden="true"
              />
              <input
                id="contest-search"
                type="search"
                value={page.query}
                onChange={(event) => page.setQuery(event.target.value)}
                placeholder="ABC 350 など"
                className={`${inputClassName} pl-9`}
              />
            </div>
            <p className="mt-3 text-[11px] leading-5 text-muted-foreground">
              プレビューではABC 350〜352のA〜D問題を表示しています。
            </p>
            <div className="mt-4 grid grid-cols-3 gap-2 xl:grid-cols-1">
              {page.visibleContests.map((id) => (
                <button
                  key={id}
                  type="button"
                  aria-pressed={id === page.activeContest}
                  onClick={() => page.setContestId(id)}
                  className={`rounded-xl border p-2.5 text-left transition-colors sm:p-3 ${id === page.activeContest ? "border-primary/20 bg-accent text-primary" : "border-transparent hover:bg-muted"}`}
                >
                  <span className="block text-xs font-semibold whitespace-nowrap sm:text-sm">
                    ABC {id.slice(3)}
                  </span>
                  <span className="mt-1 hidden text-[11px] text-muted-foreground sm:block">
                    AtCoder Beginner Contest
                  </span>
                </button>
              ))}
            </div>
            {!page.visibleContests.length && (
              <p className="py-6 text-sm text-muted-foreground">該当するコンテストはありません。</p>
            )}
          </aside>

          <section
            className="min-w-0 overflow-hidden rounded-2xl border border-border bg-card shadow-card"
            aria-labelledby="contest-problems-title"
          >
            <header className="flex flex-wrap items-center justify-between gap-3 border-b border-border p-5 sm:px-6">
              <div>
                <h2 id="contest-problems-title" className="font-semibold">
                  {page.activeContest ? `ABC ${page.activeContest.slice(3)}` : "コンテストを選択"}
                </h2>
                <p className="mt-1 text-xs text-muted-foreground">
                  復習したい問題を選択してください。未ACでも登録できます。
                </p>
              </div>
              <span className="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">
                {page.contestProblems.length}問
              </span>
            </header>
            {!!page.contestProblems.length && (
              <label className="flex min-h-11 items-center gap-3 border-b border-border bg-background/70 px-5 text-xs text-muted-foreground sm:px-6">
                <input
                  type="checkbox"
                  checked={page.allSelected}
                  disabled={!page.hasAvailable}
                  onChange={page.toggleAll}
                  className="size-4 accent-primary"
                />
                未登録の問題をすべて選択
              </label>
            )}
            <div className="divide-y divide-border">
              {page.contestProblems.map((problem) => (
                <div
                  key={problem.id}
                  className={`flex items-center gap-3 px-5 py-5 sm:gap-4 sm:px-6 ${page.selected.has(problem.id) ? "bg-accent/40" : "hover:bg-background/50"}`}
                >
                  <input
                    type="checkbox"
                    aria-label={`${problem.index} ${problem.title}を選択`}
                    checked={page.selected.has(problem.id)}
                    disabled={page.registered.has(problem.id)}
                    onChange={() => page.toggleProblem(problem.id)}
                    className="size-4 shrink-0 accent-primary"
                  />
                  <div className="min-w-0 flex-1">
                    <ProblemIdentity problem={problem} />
                  </div>
                  <span className="hidden text-xs text-muted-foreground md:block">
                    提出情報未確認
                  </span>
                  {page.registered.has(problem.id) && (
                    <span className="flex shrink-0 items-center gap-1 text-xs text-primary">
                      <Check className="size-3.5" aria-hidden="true" />
                      <span className="sr-only sm:not-sr-only">登録済み</span>
                    </span>
                  )}
                  <a
                    href={problemUrl(problem)}
                    target="_blank"
                    rel="noreferrer"
                    aria-label={`${problem.title}をAtCoderで開く`}
                    className="grid size-10 shrink-0 place-items-center rounded-lg text-muted-foreground hover:bg-accent hover:text-primary"
                  >
                    <ExternalLink className="size-4" aria-hidden="true" />
                  </a>
                </div>
              ))}
              {!page.contestProblems.length && (
                <p className="p-10 text-center text-sm text-muted-foreground">
                  検索条件を変えるか、URLから登録できます。
                </p>
              )}
            </div>
            <div className="border-t border-border bg-background/60 p-5 sm:p-6">
              <label htmlFor="registration-note" className="mb-2 block text-xs font-medium">
                登録メモ{" "}
                <span className="ml-1 text-muted-foreground">任意・選択した問題に共通</span>
              </label>
              <input
                id="registration-note"
                value={page.note}
                onChange={(event) => page.setNote(event.target.value)}
                maxLength={500}
                placeholder="例：解説の考え方を自力で再現したい"
                className={inputClassName}
              />
              <div className="mt-4 flex flex-wrap items-center justify-between gap-4">
                <p className="text-xs text-muted-foreground">
                  <span className="mr-1 text-base font-semibold text-primary">
                    {page.selectedCount}
                  </span>
                  問を選択中
                </p>
                <Button
                  className="h-11 gap-2 px-5"
                  disabled={!page.selectedCount}
                  onClick={page.registerSelected}
                >
                  <Plus aria-hidden="true" />
                  選択した問題を追加
                </Button>
              </div>
            </div>
          </section>
        </div>
      ) : (
        <div className="grid items-start gap-5 lg:grid-cols-[minmax(0,1fr)_280px]">
          <form
            noValidate
            onSubmit={page.submitUrl}
            className="min-w-0 rounded-2xl border border-border bg-card p-6 shadow-card sm:p-8"
          >
            <span className="mb-5 grid size-11 place-items-center rounded-xl bg-accent text-primary">
              <Link2 className="size-5" aria-hidden="true" />
            </span>
            <h2 className="text-lg font-semibold">問題のURLを貼り付ける</h2>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              コンテストが見つからない問題も、URLから登録できます。
            </p>
            <label htmlFor="problem-url" className="mt-6 mb-2 block text-sm font-medium">
              AtCoderの問題URL
            </label>
            <input
              id="problem-url"
              type="url"
              value={page.url}
              onChange={(event) => page.setUrl(event.target.value)}
              placeholder="https://atcoder.jp/contests/abc350/tasks/abc350_a"
              required
              className={inputClassName}
              aria-invalid={!!page.urlError}
              aria-describedby={page.urlError ? "problem-url-error" : "problem-url-hint"}
            />
            <p id="problem-url-hint" className="mt-2 text-xs leading-5 text-muted-foreground">
              URLの形式のみ確認します。サンプルにない問題は問題IDで表示し、実在確認や情報取得は行いません。
            </p>
            {page.urlError && (
              <p id="problem-url-error" role="alert" className="mt-3 text-xs leading-5 text-danger">
                {page.urlError}
              </p>
            )}
            <label htmlFor="url-note" className="mt-5 mb-2 block text-sm font-medium">
              登録メモ <span className="text-xs text-muted-foreground">任意</span>
            </label>
            <textarea
              id="url-note"
              value={page.note}
              onChange={(event) => page.setNote(event.target.value)}
              maxLength={500}
              rows={3}
              placeholder="復習したい理由をひとこと。"
              className={`${inputClassName} resize-y`}
            />
            <Button type="submit" className="mt-6 h-11 gap-2 px-5">
              <Plus aria-hidden="true" />
              復習リストに追加
            </Button>
          </form>
          <aside className="rounded-2xl border border-primary/15 bg-aurora p-6">
            <CalendarDays className="size-5 text-primary" aria-hidden="true" />
            <h2 className="mt-4 text-sm font-semibold">最初の復習は、明日。</h2>
            <p className="mt-2 text-sm font-semibold text-primary">{formatDay(page.nextDay)}</p>
            <p className="mt-3 text-xs leading-6 text-muted-foreground">
              少し間を空けて、思い出す練習を。プレビューでは翌日を初回の予定日にしています。
            </p>
            <Link
              to="/reviews"
              className="mt-5 inline-flex items-center gap-2 rounded-md text-xs font-semibold text-primary"
            >
              今日の復習を見る
              <ArrowRight className="size-3.5" aria-hidden="true" />
            </Link>
          </aside>
        </div>
      )}
    </LearningPageLayout>
  );
}
