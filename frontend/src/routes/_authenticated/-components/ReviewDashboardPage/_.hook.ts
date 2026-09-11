import { Code, ConnectError } from "@connectrpc/connect";
import { useQuery } from "@connectrpc/connect-query";
import { skipToken } from "@tanstack/react-query";
import { useRouter, useSearch } from "@tanstack/react-router";
import { ReviewService } from "@/__generated__/spaco/review/v1/review_pb";
import { connectErrorMessage } from "@/lib/connect";

export function useReviewDashboardPage() {
  const { problem: selectedProblemId } = useSearch({ from: "/_authenticated/reviews" });
  const router = useRouter();
  const todayQuery = useQuery(ReviewService.method.getTodayReviews, {});
  const selectionQuery = useQuery(
    ReviewService.method.getReviewItem,
    selectedProblemId ? { problemId: selectedProblemId } : skipToken,
  );

  const data = todayQuery.data;
  const today = data?.today ?? "";
  const due = data?.dueItems ?? [];
  const completed = data?.completedTodayCount ?? 0;
  const selectedItem = selectionQuery.data?.item;
  const selectionMissing =
    selectionQuery.isError && ConnectError.from(selectionQuery.error).code === Code.NotFound;

  return {
    loading: todayQuery.isPending,
    error: todayQuery.isError && !data ? connectErrorMessage(todayQuery.error) : undefined,
    retry: () => void todayQuery.refetch(),
    today,
    timeZone: data?.timeZone ?? "",
    due,
    completed,
    upcoming: data?.upcomingItems ?? [],
    overdue: due.filter((item) => item.dueOn < today).length,
    selected: selectedItem && !selectedItem.paused ? selectedItem : undefined,
    invalidSelection: !!selectedProblemId && (selectionMissing || !!selectedItem?.paused),
    selectionError:
      selectionQuery.isError && !selectionMissing
        ? connectErrorMessage(selectionQuery.error)
        : undefined,
    progress:
      due.length + completed === 0 ? 0 : Math.round((completed / (due.length + completed)) * 100),
    start: (problemId: string) =>
      void router.navigate({ to: "/reviews", search: { problem: problemId }, replace: true }),
    close: () => void router.navigate({ to: "/reviews", search: {}, replace: true }),
  };
}
