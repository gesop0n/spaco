import type { ReactNode } from "react";
import { useAuthVerification } from "./_.hook";

type Props = {
  children: ReactNode;
};

function AuthBoundaryErrorFallback({ onRetryClick }: { onRetryClick: () => void }) {
  return (
    <main className="grid min-h-svh place-content-center gap-4 px-6 text-center">
      <h1 className="text-xl font-semibold">接続エラーが発生しました</h1>
      <p className="text-sm leading-6 text-fg">
        ネットワーク接続を確認のうえ、もう一度お試しください。
      </p>
      <button
        type="button"
        onClick={onRetryClick}
        className="mx-auto min-h-11 rounded-lg bg-accent px-5 py-2.5 text-sm font-semibold text-surface hover:bg-accent-hover"
      >
        再試行
      </button>
    </main>
  );
}

/** 認証確認が完了するまで保護コンテンツを表示しないためのboundary。 */
export function WithAuthBoundary({ children }: Readonly<Props>) {
  const { state, onRetryClick } = useAuthVerification();

  switch (state) {
    case "ready":
      return children;
    case "transientError":
      return <AuthBoundaryErrorFallback onRetryClick={onRetryClick} />;
    case "loading":
      // 通常は一瞬で完了するため、spinnerのflashを避ける。
      return null;
  }
}
