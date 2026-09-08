import { useState } from "react";
import { useReviewWorkspace } from "../ReviewWorkspace";

export function useReviewListPage() {
  const { problems, dispatch, today, timeZone } = useReviewWorkspace();
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<"all" | "active" | "paused">("all");
  const search = query.trim().toLowerCase();
  const visible = problems.filter(
    (problem) =>
      (filter === "all" || (filter === "paused" ? problem.paused : !problem.paused)) &&
      `${problem.title} ${problem.id} ${problem.contestId}`.toLowerCase().includes(search),
  );
  const counts = {
    all: problems.length,
    active: problems.filter((problem) => !problem.paused).length,
    paused: problems.filter((problem) => problem.paused).length,
  };
  return {
    visible,
    query,
    setQuery,
    filter,
    setFilter,
    counts,
    today,
    timeZone,
    togglePause: (problemId: string) => dispatch({ type: "togglePause", problemId }),
  };
}
