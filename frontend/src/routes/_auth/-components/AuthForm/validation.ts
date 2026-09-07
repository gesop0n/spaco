import type { AuthMode } from "./types";

export function getFieldError(input: HTMLInputElement, mode: AuthMode): string | undefined {
  if (input.name === "email") {
    if (input.validity.valueMissing) return "メールアドレスを入力してください。";
    if (input.validity.typeMismatch) return "メールアドレスの形式を確認してください。";
  }

  if (input.name === "password") {
    if (!input.value) return "パスワードを入力してください。";
    if (mode === "register" && input.value.length < 8) {
      return "パスワードは8文字以上で入力してください。";
    }
  }

  return undefined;
}
