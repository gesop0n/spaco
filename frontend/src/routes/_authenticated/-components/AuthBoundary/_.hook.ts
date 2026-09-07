import { useQuery } from "@connectrpc/connect-query";
import { useLocation, useRouter } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
import { useAuth } from "@/components/AuthProvider";
import { AccountService } from "@/__generated__/spaco/account/v1/account_pb";
import { isAuthenticationError } from "@/lib/connect";

export type AuthVerificationState = "loading" | "ready" | "transientError";

/** 認証済みrouteでsessionとアプリ内accountの両方を確認する。 */
export function useAuthVerification(): {
  state: AuthVerificationState;
  onRetryClick: () => void;
} {
  const authState = useAuth();
  const router = useRouter();
  const currentHref = useLocation({ select: (location) => location.href });
  const currentPathname = useLocation({ select: (location) => location.pathname });
  const redirectStartedRef = useRef(false);

  const accountQuery = useQuery(
    AccountService.method.getCurrentAccount,
    {},
    {
      enabled: authState.status === "authenticated",
    },
  );

  const rpcRequiresLogin = accountQuery.isError && isAuthenticationError(accountQuery.error);
  const account = accountQuery.data?.account;
  const needsProfileSetup = account != null && !account.setupCompleted;

  useEffect(() => {
    const sessionMissing = authState.status === "unauthenticated";
    if (!sessionMissing && !rpcRequiresLogin) {
      redirectStartedRef.current = false;
      return;
    }
    if (redirectStartedRef.current) return;
    redirectStartedRef.current = true;

    const loginURL = `/login?redirectTo=${encodeURIComponent(currentHref)}`;
    // backendがtokenを拒否した場合は、同じtokenでredirectを繰り返さないよう端末から破棄する。
    const clearRejectedSession = rpcRequiresLogin ? authState.clearSession() : Promise.resolve();
    void clearRejectedSession
      .catch(() => undefined)
      .finally(() => router.history.replace(loginURL));
  }, [authState, currentHref, router, rpcRequiresLogin]);

  useEffect(() => {
    if (needsProfileSetup && currentPathname !== "/profile") {
      void router.navigate({ to: "/profile", replace: true });
    }
  }, [currentPathname, needsProfileSetup, router]);

  const state = ((): AuthVerificationState => {
    if (authState.status === "loading") return "loading";
    if (authState.status === "error") return "transientError";
    if (authState.status === "unauthenticated" || rpcRequiresLogin) return "loading";
    if (accountQuery.isPending) return "loading";
    if (accountQuery.isError) return "transientError";
    // 成功responseにaccountが無い場合は契約違反なので、保護画面を描画しない。
    if (account == null) return "transientError";
    if (needsProfileSetup && currentPathname !== "/profile") return "loading";
    return "ready";
  })();

  const onRetryClick = () => {
    if (authState.status === "error") {
      void authState.reloadSession();
      return;
    }
    void accountQuery.refetch();
  };

  return { state, onRetryClick };
}
