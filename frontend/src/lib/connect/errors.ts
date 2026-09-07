import { Code, ConnectError } from "@connectrpc/connect";

/** 認証をやり直す必要があるConnectRPCエラーかを判定する。 */
export function isAuthenticationError(error: unknown): boolean {
  const connectError = ConnectError.from(error);
  return connectError.code === Code.Unauthenticated || connectError.code === Code.PermissionDenied;
}

/**
 * TanStack Queryから呼ばれるretry判定。
 * 認証・入力エラーは再試行しても直らないため即座に終了し、一時的な障害だけを再試行する。
 */
export function shouldRetryConnectQuery(failureCount: number, error: unknown): boolean {
  const connectError = ConnectError.from(error);
  const retryableCodes = new Set([
    Code.Unknown,
    Code.DeadlineExceeded,
    Code.Internal,
    Code.Unavailable,
  ]);

  return failureCount < 2 && retryableCodes.has(connectError.code);
}

/** RPCの実装詳細を画面へ露出させず、ユーザー向けの文言へ変換する。 */
export function connectErrorMessage(error: unknown): string {
  const connectError = ConnectError.from(error);

  switch (connectError.code) {
    case Code.Unauthenticated:
    case Code.PermissionDenied:
      return "ログイン情報を確認できませんでした。もう一度ログインしてください。";
    case Code.InvalidArgument:
      return "入力内容を確認してください。";
    case Code.Unavailable:
    case Code.DeadlineExceeded:
      return "サーバーに接続できませんでした。時間をおいて再度お試しください。";
    default:
      return "処理中にエラーが発生しました。時間をおいて再度お試しください。";
  }
}
