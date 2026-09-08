import { zodResolver } from "@hookform/resolvers/zod";
import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { AccountService } from "@/__generated__/spaco/account/v1/account_pb";
import { connectErrorMessage } from "@/lib/connect";
import { defaultTimeZone, getTimeZoneGroups } from "./time-zones";
import { profileSchema } from "./validation";
import type { ProfileFormValues } from "./validation";

/** profile取得・更新と、更新後のaccount cache同期を担当する。 */
export function useProfileForm() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const accountQuery = useQuery(AccountService.method.getCurrentAccount, {});
  const updateProfile = useMutation(AccountService.method.updateProfile);
  const [serverError, setServerError] = useState<string>();

  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    mode: "onTouched",
    defaultValues: {
      username: "",
      atcoderId: "",
      timeZone: defaultTimeZone,
    },
  });

  const account = accountQuery.data?.account;
  useEffect(() => {
    if (account == null) return;
    form.reset({
      username: account.username ?? "",
      atcoderId: account.atcoderId ?? "",
      timeZone: account.timeZone || defaultTimeZone,
    });
  }, [account, form]);

  const handleSubmit = form.handleSubmit(async (values) => {
    setServerError(undefined);
    try {
      await updateProfile.mutateAsync(values);

      // GetCurrentAccountを再取得し、AuthBoundaryと各画面に最新profileを反映する。
      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: AccountService.method.getCurrentAccount,
          cardinality: undefined,
        }),
      });
      // 初回設定はWelcomeへ、マイページからの編集はマイページへ戻す。
      await router.navigate({ to: account?.setupCompleted ? "/mypage" : "/app", replace: true });
    } catch (error) {
      setServerError(connectErrorMessage(error));
    }
  });

  return {
    ...form,
    handleSubmit,
    serverError,
    setupCompleted: account?.setupCompleted ?? false,
    timeZoneGroups: getTimeZoneGroups(account?.timeZone),
  };
}
