import { useSearch } from "@tanstack/react-router";
import { AuthForm } from "../AuthForm";

/** login routeのsearch paramsを認証フォームへ接続する。 */
export function LoginPage() {
  const { redirectTo } = useSearch({ from: "/_auth/login" });
  return <AuthForm key="login" mode="login" redirectTo={redirectTo} />;
}
