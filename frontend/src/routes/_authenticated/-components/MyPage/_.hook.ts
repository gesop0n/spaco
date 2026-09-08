import { useQuery } from "@connectrpc/connect-query";
import { AccountService } from "@/__generated__/spaco/account/v1/account_pb";

/** マイページに表示するaccountを取得する。 */
export function useMyPage() {
  const { data } = useQuery(AccountService.method.getCurrentAccount, {});

  return {
    account: data?.account,
  };
}
