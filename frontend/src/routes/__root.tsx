import { createRootRoute, Link, Outlet } from "@tanstack/react-router";

export const Route = createRootRoute({
  component: Outlet,
  notFoundComponent: () => (
    <main className="grid min-h-svh place-content-center gap-6 p-8 text-center">
      <h1 className="text-2xl leading-normal font-bold">ページが見つかりませんでした。</h1>
      <Link className="text-accent underline underline-offset-4" to="/">
        トップに戻る
      </Link>
    </main>
  ),
});
