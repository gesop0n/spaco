import { useContext } from "react";
import { AuthContext } from "./context";

/** AuthProviderが管理しているSupabaseの認証状態を取得する。 */
export function useAuth() {
  const value = useContext(AuthContext);
  if (value == null) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return value;
}
