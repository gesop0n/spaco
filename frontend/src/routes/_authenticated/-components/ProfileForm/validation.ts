import { z } from "zod";

export const profileSchema = z.object({
  username: z
    .string()
    .trim()
    .min(1, "ユーザー名を入力してください。")
    .refine((value) => Array.from(value).length <= 40, "ユーザー名は40文字以内で入力してください。")
    .regex(/^\P{Cc}*$/u, "ユーザー名に改行や制御文字は使えません。"),
  atcoderId: z
    .string()
    .trim()
    .refine((value) => new TextEncoder().encode(value).length <= 64, "AtCoder IDが長すぎます。")
    .refine((value) => !/[ \t\r\n]/.test(value), "AtCoder IDに空白は使えません。"),
  timeZone: z.string().trim().min(1, "タイムゾーンを選択してください。"),
});

export type ProfileFormValues = z.infer<typeof profileSchema>;
