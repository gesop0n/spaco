import { useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { ReviewService } from "@/__generated__/spaco/review/v1/review_pb";
import type { ReviewItem, ReviewLog } from "@/__generated__/spaco/review/v1/review_pb";
import { connectErrorMessage } from "@/lib/connect";
import { reviewServiceQueryKey } from "../ReviewWorkspace/api";

export function useReviewListPage() {
  const queryClient = useQueryClient();
  const listQuery = useQuery(ReviewService.method.listReviewItems, {});
  const pauseReviewItem = useMutation(ReviewService.method.pauseReviewItem);
  const resumeReviewItem = useMutation(ReviewService.method.resumeReviewItem);
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<"all" | "active" | "paused">("all");
  const [pendingItemId, setPendingItemId] = useState<string>();
  const [actionError, setActionError] = useState<string>();

  const data = listQuery.data;
  const items = data?.items ?? [];
  const logsByItem = useMemo(() => {
    const grouped = new Map<string, ReviewLog[]>();
    for (const log of data?.logs ?? []) {
      const logs = grouped.get(log.reviewItemId);
      if (logs) logs.push(log);
      else grouped.set(log.reviewItemId, [log]);
    }
    return grouped;
  }, [data]);

  const search = query.trim().toLowerCase();
  const visible = items.filter((item) => {
    if (filter !== "all" && (filter === "paused") !== item.paused) return false;
    const problem = item.problem;
    return `${problem?.name ?? ""} ${problem?.id ?? ""} ${problem?.contestId ?? ""}`
      .toLowerCase()
      .includes(search);
  });

  async function togglePause(item: ReviewItem) {
    if (pendingItemId) return;
    setPendingItemId(item.id);
    setActionError(undefined);
    try {
      const request = { reviewItemId: item.id };
      await (item.paused
        ? resumeReviewItem.mutateAsync(request)
        : pauseReviewItem.mutateAsync(request));
      await queryClient.invalidateQueries({ queryKey: reviewServiceQueryKey });
    } catch (error) {
      setActionError(connectErrorMessage(error));
    } finally {
      setPendingItemId(undefined);
    }
  }

  return {
    loading: listQuery.isPending,
    error: listQuery.isError && !data ? connectErrorMessage(listQuery.error) : undefined,
    retry: () => void listQuery.refetch(),
    visible,
    query,
    setQuery,
    filter,
    setFilter,
    counts: {
      all: items.length,
      active: items.filter((item) => !item.paused).length,
      paused: items.filter((item) => item.paused).length,
    },
    today: data?.today ?? "",
    timeZone: data?.timeZone ?? "",
    logsOf: (item: ReviewItem) => logsByItem.get(item.id) ?? [],
    pendingItemId,
    actionError,
    togglePause: (item: ReviewItem) => void togglePause(item),
  };
}
