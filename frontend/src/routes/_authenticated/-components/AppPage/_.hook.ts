import { useQuery } from "@connectrpc/connect-query";
import { useRouter } from "@tanstack/react-router";
import { useState } from "react";
import { useAuth } from "@/components/AuthProvider";
import { AccountService } from "@/__generated__/spaco/account/v1/account_pb";

/** マイページに必要なaccount取得とログアウト処理をまとめる。 */
export function useAppPage() {
  const router = useRouter();
  const { signOut } = useAuth();
  const { data } = useQuery(AccountService.method.getCurrentAccount, {});
  const [signOutError, setSignOutError] = useState<string>();

  const handleSignOut = async () => {
    setSignOutError(undefined);
    try {
      await signOut();
      await router.navigate({ to: "/", replace: true });
    } catch {
      setSignOutError("ログアウトに失敗しました。時間をおいて再度お試しください。");
    }
  };

  return {
    account: data?.account,
    handleSignOut,
    signOutError,
  };
}
