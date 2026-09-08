import { createFileRoute, Navigate } from "@tanstack/react-router";
import { useAuth } from "@/components/AuthProvider";
import { LandingPage } from "./-components/LandingPage";

export const Route = createFileRoute("/")({
  component: function IndexPage() {
    const authState = useAuth();

    // メール確認URLからのsession復元を待ち、未確認の状態で公開ページを表示しない。
    if (authState.status === "loading") return null;

    if (authState.status === "authenticated") {
      // プロフィール設定の要否は、保護routeのAuthBoundaryで一元的に判定する。
      return <Navigate to="/app" replace />;
    }

    if (authState.status === "error") {
      return (
        <main className="grid min-h-svh place-content-center gap-4 px-6 text-center">
          <h1 className="text-xl font-semibold">ログイン状態を確認できませんでした</h1>
          <p className="text-sm leading-6 text-fg">接続を確認して、もう一度お試しください。</p>
          <button
            type="button"
            onClick={() => void authState.reloadSession()}
            className="mx-auto min-h-11 rounded-lg bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground hover:bg-accent-hover"
          >
            再試行
          </button>
        </main>
      );
    }

    return <LandingPage />;
  },
});
