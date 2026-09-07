import type { Session } from "@supabase/auth-js";
import { createContext } from "react";

/** Supabase sessionの初期読み込みを含む、クライアント側の認証状態。 */
export type AuthState =
  | { status: "loading"; session: null; error: null }
  | { status: "authenticated"; session: Session; error: null }
  | { status: "unauthenticated"; session: null; error: null }
  | { status: "error"; session: null; error: Error };

export type AuthContextValue = AuthState & {
  /** Supabase側でもsessionを無効化する通常のログアウト。 */
  signOut: () => Promise<void>;
  /** backendがtokenを拒否した場合などに、端末上のsessionだけを確実に破棄する。 */
  clearSession: () => Promise<void>;
  /** 初期読み込みに失敗した認証状態を再取得する。 */
  reloadSession: () => Promise<void>;
};

export const AuthContext = createContext<AuthContextValue | undefined>(undefined);
