import { env } from "@/env";
import { AuthClient } from "@supabase/auth-js";

export const auth = new AuthClient({
  url: `${env.VITE_SUPABASE_URL}/auth/v1`,

  headers: {
    apikey: env.VITE_SUPABASE_PUBLISHABLE_KEY,
    Authorization: `Bearer ${env.VITE_SUPABASE_PUBLISHABLE_KEY}`,
  },

  // アクセストークンの自動更新
  autoRefreshToken: true,

  // localStorage に保存
  persistSession: true,

  // OAuth / マジックリンクのコールバック処理
  detectSessionInUrl: true,

  flowType: "pkce",
  storageKey: "spaco-auth",
});
