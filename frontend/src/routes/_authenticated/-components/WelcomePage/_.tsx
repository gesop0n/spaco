import { Link } from "@tanstack/react-router";
import { ArrowUpRight, BookOpen, CalendarCheck2, RotateCcw, Search } from "lucide-react";

const features = [
  {
    icon: CalendarCheck2,
    title: "今日の復習",
    description: "今日取り組む問題を、ひと目で。",
    to: "/reviews",
  },
  {
    icon: Search,
    title: "問題を登録",
    description: "もう一度解きたい問題を見つけよう。",
    to: "/problems",
  },
  {
    icon: BookOpen,
    title: "復習リスト",
    description: "積み重ねた記録と、これからの予定。",
    to: "/review-list",
  },
] as const;

/** 学習機能が揃うまでの、認証後の仮のホーム画面。 */
export function WelcomePage() {
  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col justify-center gap-8 px-5 py-12 md:px-10 md:py-16">
      <section className="relative isolate overflow-hidden rounded-3xl border border-border bg-card p-7 shadow-card sm:p-12">
        <div className="pointer-events-none absolute inset-0 -z-1 bg-aurora opacity-50" />
        <span className="mb-8 grid size-14 place-items-center rounded-2xl border border-card bg-card/80 text-primary shadow-action">
          <RotateCcw className="size-6" aria-hidden="true" />
        </span>
        <p className="mb-3 text-xs font-semibold tracking-[0.2em] text-primary">WELCOME</p>
        <h1 className="text-3xl leading-snug font-semibold tracking-tight sm:text-4xl">
          spacoへようこそ。
        </h1>
        <p className="mt-4 text-base leading-8 text-foreground">
          解いた問題を、次の「解ける」につなげよう。
        </p>
        <p className="mt-2 max-w-lg text-sm leading-7 text-muted-foreground">
          あなたのペースで、少しずつ。まずはサンプルデータで、問題の登録と復習を試してみましょう。操作内容は再読み込みでリセットされます。
        </p>
      </section>

      <section aria-labelledby="upcoming-title">
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <h2 id="upcoming-title" className="text-sm font-semibold">
            学習スペースを試す
          </h2>
          <span className="rounded-md border border-border bg-card/70 px-2 py-0.5 text-[11px] text-muted-foreground">
            UIプレビュー
          </span>
        </div>
        <div className="grid gap-3 lg:grid-cols-3">
          {features.map(({ icon: Icon, title, description, to }) => (
            <Link
              key={title}
              to={to}
              className="group rounded-2xl border border-border bg-card/80 p-5 transition-colors hover:border-primary/30 hover:bg-accent/60"
            >
              <div className="mb-4 flex items-center justify-between">
                <Icon className="size-5 text-primary" aria-hidden="true" />
                <ArrowUpRight
                  className="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5"
                  aria-hidden="true"
                />
              </div>
              <h3 className="text-sm font-semibold">{title}</h3>
              <p className="mt-2 text-xs leading-6 text-muted-foreground">{description}</p>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}
