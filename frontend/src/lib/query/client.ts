import { QueryClient } from "@tanstack/react-query";
import { shouldRetryConnectQuery } from "@/lib/connect";

/**
 * QueryClientはrenderのたびに作り直さず、ブラウザ上で一つだけ共有する。
 * 共通retry方針はここへ集約し、個別queryで必要な場合だけ上書きする。
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: shouldRetryConnectQuery,
      staleTime: 30_000,
    },
    mutations: {
      retry: false,
    },
  },
});
