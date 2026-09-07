import { useEffect, useRef, useState } from "react";
import type { ChangeEvent, FocusEvent, SubmitEvent } from "react";
import type { AuthField, AuthMode, FieldErrors } from "./types";
import { getFieldError } from "./validation";

export function useAuthForm(mode: AuthMode) {
  const emailRef = useRef<HTMLInputElement>(null);
  const passwordRef = useRef<HTMLInputElement>(null);
  const recoveryDialogRef = useRef<HTMLDialogElement>(null);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [showAvailability, setShowAvailability] = useState(false);

  useEffect(() => {
    const previousTitle = document.title;
    document.title = `${mode === "register" ? "新規登録" : "ログイン"} | spaco`;
    return () => {
      document.title = previousTitle;
    };
  }, [mode]);

  function handleBlur(event: FocusEvent<HTMLInputElement>) {
    const input = event.currentTarget;
    setErrors((current) => ({ ...current, [input.name]: getFieldError(input, mode) }));
  }

  function handleChange(event: ChangeEvent<HTMLInputElement>) {
    const input = event.currentTarget;
    const field = input.name as AuthField;
    if (errors[field]) {
      setErrors((current) => ({ ...current, [field]: getFieldError(input, mode) }));
    }
    setShowAvailability(false);
  }

  function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    const email = emailRef.current;
    const password = passwordRef.current;
    if (!email || !password) return;

    const nextErrors = {
      email: getFieldError(email, mode),
      password: getFieldError(password, mode),
    };
    setErrors(nextErrors);
    setShowAvailability(false);

    if (nextErrors.email) {
      email.focus();
    } else if (nextErrors.password) {
      password.focus();
    } else {
      // Replace this notice with the authentication request when the API is available.
      setShowAvailability(true);
    }
  }

  return {
    emailRef,
    passwordRef,
    recoveryDialogRef,
    errors,
    passwordVisible,
    showAvailability,
    handleBlur,
    handleChange,
    handleSubmit,
    togglePasswordVisibility: () => setPasswordVisible((visible) => !visible),
    openRecoveryDialog: () => recoveryDialogRef.current?.showModal(),
    closeRecoveryDialog: () => recoveryDialogRef.current?.close(),
  };
}
