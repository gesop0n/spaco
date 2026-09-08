import { useQuery } from "@connectrpc/connect-query";
import { useReducer } from "react";
import type { ReactNode } from "react";
import { useAuth } from "@/components/AuthProvider";
import { AccountService } from "@/__generated__/spaco/account/v1/account_pb";
import { ReviewWorkspaceContext } from "./context";
import { sampleProblems } from "./fixtures";
import { localDateTime, studyReducer } from "./model";

function PreviewState({ children, timeZone }: { children: ReactNode; timeZone: string }) {
  const today = localDateTime(timeZone).slice(0, 10);
  const [problems, dispatch] = useReducer(studyReducer, today, sampleProblems);
  return (
    <ReviewWorkspaceContext.Provider value={{ problems, dispatch, today, timeZone }}>
      {children}
    </ReviewWorkspaceContext.Provider>
  );
}

/** API未接続のUIプレビュー。データはメモリ内に限り、アカウント切替時も分離する。 */
export function ReviewWorkspaceProvider({ children }: { children: ReactNode }) {
  const auth = useAuth();
  const { data } = useQuery(AccountService.method.getCurrentAccount, {});
  return (
    <PreviewState key={auth.session?.user.id} timeZone={data?.account?.timeZone || "Asia/Tokyo"}>
      {children}
    </PreviewState>
  );
}
