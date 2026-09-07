import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { useAuth } from "@/components/AuthProvider";
import { auth } from "@/lib/supabase";
import type { AuthMode } from "./types";
import { authSchemas } from "./validation";
import type { AuthFormValues } from "./validation";

/** 外部URLへのredirectを防ぎ、認証後の遷移先をアプリ内pathだけに制限する。 */
function safeRedirectPath(redirectTo: string | undefined): string {
  if (
    redirectTo == null ||
    redirectTo === "/" ||
    !redirectTo.startsWith("/") ||
    redirectTo.startsWith("//") ||
    redirectTo.startsWith("/login") ||
    redirectTo.startsWith("/register")
  ) {
    return "/app";
  }
  return redirectTo;
}

function authenticationErrorMessage(errorCode: string | undefined): string {
  switch (errorCode) {
    case "invalid_credentials":
      return "メールアドレスまたはパスワードが正しくありません。";
    case "email_not_confirmed":
      return "確認メール内のリンクを開いてからログインしてください。";
    case "user_already_exists":
      return "このメールアドレスはすでに登録されています。";
    case "over_email_send_rate_limit":
    case "over_request_rate_limit":
      return "操作回数が上限に達しました。時間をおいて再度お試しください。";
    default:
      return "認証処理に失敗しました。時間をおいて再度お試しください。";
  }
}

export function useAuthForm(mode: AuthMode, redirectTo?: string) {
  const recoveryDialogRef = useRef<HTMLDialogElement>(null);
  const router = useRouter();
  const authState = useAuth();
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [formError, setFormError] = useState<string>();
  const [notice, setNotice] = useState<string>();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<AuthFormValues>({
    resolver: zodResolver(authSchemas[mode]),
    mode: "onTouched",
    defaultValues: { email: "", password: "" },
  });

  useEffect(() => {
    const previousTitle = document.title;
    document.title = `${mode === "register" ? "新規登録" : "ログイン"} | spaco`;
    return () => {
      document.title = previousTitle;
    };
  }, [mode]);

  const destination = safeRedirectPath(redirectTo);

  useEffect(() => {
    // ログイン済みのユーザーが認証画面へ戻った場合も、保護画面へ送り返す。
    if (authState.status === "authenticated") {
      void router.navigate({ href: destination, replace: true });
    }
  }, [authState.status, destination, router]);

  const submit = handleSubmit(async ({ email, password }) => {
    setFormError(undefined);
    setNotice(undefined);

    try {
      if (mode === "login") {
        const { error } = await auth.signInWithPassword({ email, password });
        if (error) {
          setFormError(authenticationErrorMessage(error.code));
          return;
        }

        // Supabaseへのログイン成功直後はonAuthStateChangeの反映より先にnavigateが
        // 実行される場合がある。保護routeが古いunauthenticated状態を見ないよう、
        // Contextのsessionを同期してから遷移する。
        await authState.reloadSession();
        await router.navigate({ href: destination, replace: true });
        return;
      }

      const { data, error } = await auth.signUp({ email, password });
      if (error) {
        setFormError(authenticationErrorMessage(error.code));
        return;
      }

      if (data.session == null) {
        setNotice("確認メールを送信しました。メール内のリンクを開いて登録を完了してください。");
        return;
      }

      // メール確認不要の設定ではsignUp時点でsessionが発行されるため、loginと同様に同期する。
      await authState.reloadSession();
      await router.navigate({ href: destination, replace: true });
    } catch {
      setFormError("認証サーバーに接続できませんでした。時間をおいて再度お試しください。");
    }
  });

  return {
    recoveryDialogRef,
    errors,
    formError,
    isSubmitting,
    notice,
    passwordVisible,
    emailField: register("email"),
    passwordField: register("password"),
    handleSubmit: submit,
    hideFeedback: () => {
      setFormError(undefined);
      setNotice(undefined);
    },
    togglePasswordVisibility: () => setPasswordVisible((visible) => !visible),
    openRecoveryDialog: () => recoveryDialogRef.current?.showModal(),
    closeRecoveryDialog: () => recoveryDialogRef.current?.close(),
  };
}
