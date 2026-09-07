import { z } from "zod";

export const profileSchema = z.object({
  atcoderId: z.string().trim().min(1, "AtCoder IDを入力してください。"),
  timeZone: z.string().trim().min(1, "タイムゾーンを入力してください。"),
});

export type ProfileFormValues = z.infer<typeof profileSchema>;
