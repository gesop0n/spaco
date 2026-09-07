import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useRef, useState } from "react";
import { useForm } from "react-hook-form";
import type { AuthMode } from "./types";
import { authSchemas } from "./validation";
import type { AuthFormValues } from "./validation";

export function useAuthForm(mode: AuthMode) {
  const recoveryDialogRef = useRef<HTMLDialogElement>(null);
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [showAvailability, setShowAvailability] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
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

  return {
    recoveryDialogRef,
    errors,
    passwordVisible,
    showAvailability,
    emailField: register("email"),
    passwordField: register("password"),
    // Replace this notice with the authentication request when the API is available.
    handleSubmit: handleSubmit(() => setShowAvailability(true)),
    hideAvailability: () => setShowAvailability(false),
    togglePasswordVisibility: () => setPasswordVisible((visible) => !visible),
    openRecoveryDialog: () => recoveryDialogRef.current?.showModal(),
    closeRecoveryDialog: () => recoveryDialogRef.current?.close(),
  };
}
