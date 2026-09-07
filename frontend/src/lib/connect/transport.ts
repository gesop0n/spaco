import { Code, ConnectError } from "@connectrpc/connect";
import type { Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { env } from "@/env";
import { auth } from "@/lib/supabase";

/**
 * RPC送信直前に最新のSupabase sessionを取得し、access tokenをBearer tokenとして付ける。
 * tokenの更新期限はSupabase Authに任せ、Connect側ではtokenを保存しない。
 */
const authenticationInterceptor: Interceptor = (next) => async (request) => {
  try {
    const { data, error } = await auth.getSession();
    if (error) throw error;

    if (data.session != null) {
      request.header.set("Authorization", `Bearer ${data.session.access_token}`);
    }
  } catch (error) {
    throw new ConnectError(
      "failed to get authentication session",
      Code.Unauthenticated,
      undefined,
      undefined,
      error,
    );
  }

  return next(request);
};

/** アプリ全体で共有するConnect Protocol用transport。 */
export const transport = createConnectTransport({
  baseUrl: env.VITE_API_BASE_URL.replace(/\/$/, ""),
  interceptors: [authenticationInterceptor],
});
