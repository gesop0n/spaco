import type { Session } from "@supabase/auth-js";
import { useQueryClient } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { auth } from "@/lib/supabase";
import { AuthContext } from "./context";
import type { AuthContextValue, AuthState } from "./context";

type Props = {
  children: ReactNode;
};

const initialState: AuthState = { status: "loading", session: null, error: null };

function asError(error: unknown): Error {
  return error instanceof Error ? error : new Error("failed to load authentication session");
}

/**
 * Supabase sessionをアプリ全体へ提供する。
 * APIデータはここへ保持せず、TanStack Queryのcacheへ責務を分離する。
 */
export function AuthProvider({ children }: Readonly<Props>) {
  const queryClient = useQueryClient();
  const [state, setState] = useState<AuthState>(initialState);
  const previousUserIDRef = useRef<string | null | undefined>(undefined);

  const applySession = useCallback(
    (session: Session | null) => {
      const nextUserID = session?.user.id ?? null;
      const previousUserID = previousUserIDRef.current;

      // ログアウトや別ユーザーへの切替時に、以前のユーザーのAPI cacheを残さない。
      if (previousUserID !== undefined && previousUserID !== nextUserID) {
        queryClient.clear();
      }
      previousUserIDRef.current = nextUserID;

      setState(
        session == null
          ? { status: "unauthenticated", session: null, error: null }
          : { status: "authenticated", session, error: null },
      );
    },
    [queryClient],
  );

  const reloadSession = useCallback(async () => {
    setState(initialState);
    try {
      const { data, error } = await auth.getSession();
      if (error) {
        setState({ status: "error", session: null, error });
        return;
      }
      applySession(data.session);
    } catch (error) {
      // 想定外のstorage例外などでもloadingのまま止めず、再試行UIへ遷移させる。
      setState({ status: "error", session: null, error: asError(error) });
    }
  }, [applySession]);

  useEffect(() => {
    let active = true;
    let authEventCount = 0;

    // TOKEN_REFRESHEDもここへ届くため、Contextのsessionは常に最新tokenへ更新される。
    const { data } = auth.onAuthStateChange((_event, session) => {
      authEventCount += 1;
      if (active) applySession(session);
    });

    const eventCountAtRequest = authEventCount;
    void auth
      .getSession()
      .then(({ data: sessionData, error }) => {
        // getSession中により新しいauth eventが届いた場合は、古い結果で上書きしない。
        if (!active || authEventCount !== eventCountAtRequest) return;
        if (error) {
          setState({ status: "error", session: null, error });
          return;
        }
        applySession(sessionData.session);
      })
      .catch((error: unknown) => {
        if (active && authEventCount === eventCountAtRequest) {
          setState({ status: "error", session: null, error: asError(error) });
        }
      });

    return () => {
      active = false;
      data.subscription.unsubscribe();
    };
  }, [applySession]);

  const performSignOut = useCallback(
    async (scope: "global" | "local") => {
      const { error } = await auth.signOut({ scope });
      if (error) throw error;
      applySession(null);
    },
    [applySession],
  );

  const signOut = useCallback(() => performSignOut("global"), [performSignOut]);
  const clearSession = useCallback(() => performSignOut("local"), [performSignOut]);

  const value = useMemo<AuthContextValue>(
    () => ({ ...state, signOut, clearSession, reloadSession }),
    [clearSession, reloadSession, signOut, state],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
