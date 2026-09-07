import { Link } from "@tanstack/react-router";
import { useAppPage } from "./_.hook";

/** 認証連携を確認できる最小の保護ページ。機能追加時は各業務画面へ置き換える。 */
export function AppPage() {
  const { account, handleSignOut, signOutError } = useAppPage();

  return (
    <main className="mx-auto flex min-h-svh w-full max-w-3xl flex-col gap-8 px-6 py-12">
      <header className="flex items-center justify-between gap-4">
        <Link to="/" className="font-brand text-3xl font-bold tracking-[-0.06em]">
          spaco<span className="text-accent">.</span>
        </Link>
        <button
          type="button"
          onClick={handleSignOut}
          className="rounded-lg px-4 py-2 text-sm font-semibold text-accent hover:bg-accent-soft"
        >
          ログアウト
        </button>
      </header>

      <section className="rounded-2xl border border-border bg-surface p-8 shadow-card">
        <h1 className="text-2xl font-bold">マイページ</h1>
        <dl className="mt-6 grid grid-cols-[auto_1fr] gap-x-6 gap-y-3 text-sm">
          <dt className="text-fg">ユーザーID</dt>
          <dd className="break-all">{account?.id}</dd>
          <dt className="text-fg">AtCoder ID</dt>
          <dd>{account?.atcoderId ?? "未設定"}</dd>
          <dt className="text-fg">タイムゾーン</dt>
          <dd>{account?.timeZone || "未設定"}</dd>
        </dl>
        <Link
          to="/profile"
          className="mt-7 inline-flex min-h-11 items-center rounded-lg bg-accent px-5 py-2.5 text-sm font-semibold text-surface hover:bg-accent-hover"
        >
          プロフィールを編集
        </Link>
        {signOutError && (
          <p className="mt-5 text-sm text-danger" role="alert">
            {signOutError}
          </p>
        )}
      </section>
    </main>
  );
}
