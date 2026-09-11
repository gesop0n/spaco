import { Code, ConnectError } from "@connectrpc/connect";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import { keepPreviousData, skipToken, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { AccountService } from "@/__generated__/spaco/account/v1/account_pb";
import { CatalogService } from "@/__generated__/spaco/catalog/v1/catalog_pb";
import { ReviewService } from "@/__generated__/spaco/review/v1/review_pb";
import type { ReviewItem } from "@/__generated__/spaco/review/v1/review_pb";
import { connectErrorMessage } from "@/lib/connect";
import { reviewServiceQueryKey } from "../ReviewWorkspace/api";
import { addDays, localDateTime, parseProblemUrl } from "../ReviewWorkspace/model";

const urlGuidance = "https://atcoder.jp/contests/…/tasks/… の問題URLを入力してください。";

/** ListRegisteredProblemIdsへ一度に渡せる問題数の上限。 */
const maxRegistrationCheckCount = 200;

type Notice = {
  count: number;
  alreadyRegisteredCount: number;
  firstProblemId?: string;
  firstDueOn?: string;
};

function useDebouncedValue<Value>(value: Value, delay: number): Value {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);
  return debounced;
}

export function useProblemRegistrationPage() {
  const queryClient = useQueryClient();
  const [mode, setMode] = useState<"contest" | "url">("contest");
  const [query, setQuery] = useState("");
  const [contestId, setContestId] = useState<string>();
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [note, setNote] = useState("");
  const [url, setUrl] = useState("");
  const [urlError, setUrlError] = useState<{ message: string; alreadyRegistered: boolean }>();
  const [registrationError, setRegistrationError] = useState<string>();
  const [notice, setNotice] = useState<Notice>();
  const debouncedQuery = useDebouncedValue(query.trim(), 250);

  const accountQuery = useQuery(AccountService.method.getCurrentAccount, {});
  const timeZone = accountQuery.data?.account?.timeZone || "Asia/Tokyo";

  const contestsQuery = useQuery(
    CatalogService.method.searchContests,
    { query: debouncedQuery },
    { placeholderData: keepPreviousData },
  );
  const contests = contestsQuery.data?.contests ?? [];
  const activeContestId = contests.some((contest) => contest.id === contestId)
    ? contestId
    : contests[0]?.id;

  const problemsQuery = useQuery(
    CatalogService.method.listContestProblems,
    activeContestId ? { contestId: activeContestId } : skipToken,
  );
  const activeContest =
    problemsQuery.data?.contest ?? contests.find((contest) => contest.id === activeContestId);
  const contestProblems = problemsQuery.data?.problems ?? [];

  const problemIds = contestProblems
    .slice(0, maxRegistrationCheckCount)
    .map((problem) => problem.id);
  const registeredQuery = useQuery(
    ReviewService.method.listRegisteredProblemIds,
    problemIds.length ? { problemIds } : skipToken,
  );
  const registered = new Set(registeredQuery.data?.problemIds);
  // 登録状況を確認できるまでは、登録済みの問題を誤って選べないよう選択を止める。
  const registrationChecked = registeredQuery.isSuccess;
  const available = registrationChecked
    ? contestProblems.filter((problem) => !registered.has(problem.id))
    : [];
  const selectedProblems = available.filter((problem) => selected.has(problem.id));

  const registerProblems = useMutation(ReviewService.method.registerProblems);
  const registerProblemByUrl = useMutation(ReviewService.method.registerProblemByUrl);

  function showNotice(items: ReviewItem[], alreadyRegisteredCount: number) {
    const first = items[0];
    setNotice({
      count: items.length,
      alreadyRegisteredCount,
      firstProblemId: first?.problem?.id,
      firstDueOn: first?.dueOn,
    });
  }

  async function registerSelected() {
    if (!selectedProblems.length || registerProblems.isPending) return;
    setRegistrationError(undefined);
    setNotice(undefined);
    try {
      const response = await registerProblems.mutateAsync({
        problemIds: selectedProblems.map((problem) => problem.id),
        note: note.trim(),
      });
      showNotice(response.registeredItems, response.alreadyRegisteredItems.length);
      setSelected(new Set());
      setNote("");
      await queryClient.invalidateQueries({ queryKey: reviewServiceQueryKey });
    } catch (error) {
      setRegistrationError(connectErrorMessage(error));
    }
  }

  async function submitUrl(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (registerProblemByUrl.isPending) return;
    setUrlError(undefined);
    setNotice(undefined);
    if (!parseProblemUrl(url)) {
      setUrlError({ message: urlGuidance, alreadyRegistered: false });
      return;
    }
    try {
      const response = await registerProblemByUrl.mutateAsync({
        url: url.trim(),
        note: note.trim(),
      });
      if (response.alreadyRegistered) {
        setUrlError({ message: "この問題は登録済みです。", alreadyRegistered: true });
        return;
      }
      showNotice(response.item ? [response.item] : [], 0);
      setUrl("");
      setNote("");
      await queryClient.invalidateQueries({ queryKey: reviewServiceQueryKey });
    } catch (error) {
      const invalidUrl = ConnectError.from(error).code === Code.InvalidArgument;
      setUrlError({
        message: invalidUrl ? urlGuidance : connectErrorMessage(error),
        alreadyRegistered: false,
      });
    }
  }

  function toggleProblem(id: string) {
    setSelected((previous) => {
      const next = new Set(previous);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function toggleAll() {
    setSelected((previous) => {
      const next = new Set(previous);
      const allSelected = available.every((problem) => previous.has(problem.id));
      for (const problem of available) {
        if (allSelected) next.delete(problem.id);
        else next.add(problem.id);
      }
      return next;
    });
  }

  return {
    mode,
    setMode,
    query,
    setQuery,
    selectContest: (id: string) => {
      setContestId(id);
      setSelected(new Set());
    },
    contests,
    contestsLoading: contestsQuery.isPending,
    contestsError:
      contestsQuery.isError && !contestsQuery.data
        ? connectErrorMessage(contestsQuery.error)
        : undefined,
    retryContests: () => void contestsQuery.refetch(),
    searching: query.trim() !== debouncedQuery || contestsQuery.isFetching,
    activeContestId,
    activeContest,
    contestProblems,
    problemsLoading: contestsQuery.isPending || (!!activeContestId && problemsQuery.isPending),
    problemsError: problemsQuery.isError ? connectErrorMessage(problemsQuery.error) : undefined,
    retryProblems: () => void problemsQuery.refetch(),
    registered,
    registrationChecked,
    selected,
    selectedCount: selectedProblems.length,
    allSelected: available.length > 0 && available.every((problem) => selected.has(problem.id)),
    hasAvailable: available.length > 0,
    toggleProblem,
    toggleAll,
    note,
    setNote,
    url,
    setUrl,
    urlError,
    registrationError,
    notice,
    registering: registerProblems.isPending,
    registeringUrl: registerProblemByUrl.isPending,
    nextDay: addDays(localDateTime(timeZone).slice(0, 10), 1),
    submitUrl: (event: FormEvent<HTMLFormElement>) => void submitUrl(event),
    registerSelected: () => void registerSelected(),
  };
}
