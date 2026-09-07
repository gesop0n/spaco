import { z } from "zod";
import type { AuthMode } from "./types";

const emailSchema = z
  .string()
  .trim()
  .min(1, "メールアドレスを入力してください。")
  .pipe(z.email("メールアドレスの形式を確認してください。"));

const passwordSchema = z.string().min(1, "パスワードを入力してください。");

export const loginSchema = z.object({
  email: emailSchema,
  password: passwordSchema,
});

export const registerSchema = z.object({
  email: emailSchema,
  password: passwordSchema.pipe(z.string().min(8, "パスワードは8文字以上で入力してください。")),
});

export type AuthFormValues = z.infer<typeof loginSchema>;

export const authSchemas: Record<AuthMode, z.ZodType<AuthFormValues, AuthFormValues>> = {
  login: loginSchema,
  register: registerSchema,
};
