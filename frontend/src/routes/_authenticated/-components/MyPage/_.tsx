import { Link } from "@tanstack/react-router";
import { ArrowUpRight, UserRound } from "lucide-react";
import { useMyPage } from "./_.hook";

export function MyPage() {
  const { account } = useMyPage();

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-8 px-5 py-10 md:px-10 md:py-14">
      <header>
        <p className="mb-3 text-xs font-medium tracking-[0.16em] text-primary">MY SPACE</p>
        <h1 className="text-3xl font-semibold tracking-tight">マイページ</h1>
      </header>

      <section
        className="overflow-hidden rounded-2xl border border-border bg-card shadow-card"
        aria-labelledby="account-title"
      >
        <div className="flex items-center gap-4 border-b border-border bg-aurora px-6 py-6 sm:px-8">
          <span className="grid size-11 place-items-center rounded-xl border border-card/80 bg-card/70 text-primary">
            <UserRound className="size-5" aria-hidden="true" />
          </span>
          <div>
            <h2 id="account-title" className="font-semibold">
              あなたのプロフィール
            </h2>
            <p className="mt-1 text-xs text-muted-foreground">学習に使うアカウント情報</p>
          </div>
        </div>
        <div className="p-6 sm:p-8">
          <dl className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-6 gap-y-5 text-sm">
            <dt className="text-muted-foreground">ユーザー名</dt>
            <dd className="break-all">{account?.username ?? "未設定"}</dd>
            <dt className="text-muted-foreground">AtCoder ID</dt>
            <dd className="break-all">{account?.atcoderId ?? "未設定"}</dd>
            <dt className="text-muted-foreground">タイムゾーン</dt>
            <dd className="break-all">{account?.timeZone || "未設定"}</dd>
          </dl>
          <Link
            to="/profile"
            className="mt-8 inline-flex min-h-11 items-center gap-3 rounded-lg bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground shadow-action hover:bg-accent-hover"
          >
            プロフィールを編集
            <ArrowUpRight className="size-4" aria-hidden="true" />
          </Link>
        </div>
      </section>
    </div>
  );
}
