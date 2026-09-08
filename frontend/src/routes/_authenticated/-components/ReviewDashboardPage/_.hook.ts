import { useRouter, useSearch } from "@tanstack/react-router";
import { useReviewWorkspace } from "../ReviewWorkspace";
import { dueProblems } from "../ReviewWorkspace/model";

export function useReviewDashboardPage() {
  const { problems, today, timeZone } = useReviewWorkspace();
  const { problem: selectedId } = useSearch({ from: "/_authenticated/reviews" });
  const router = useRouter();
  const due = dueProblems(problems, today);
  const completed = problems.reduce(
    (count, problem) =>
      count + problem.history.filter((entry) => entry.performedAt.slice(0, 10) === today).length,
    0,
  );
  const upcoming = problems
    .filter((problem) => !problem.paused && problem.dueOn > today)
    .sort((a, b) => a.dueOn.localeCompare(b.dueOn))
    .slice(0, 3);
  const selected = problems.find((problem) => problem.id === selectedId && !problem.paused);
  return {
    today,
    timeZone,
    due,
    completed,
    upcoming,
    selected,
    overdue: due.filter((problem) => problem.dueOn < today).length,
    invalidSelection: !!selectedId && !selected,
    progress:
      due.length + completed === 0 ? 0 : Math.round((completed / (due.length + completed)) * 100),
    start: (id: string) =>
      void router.navigate({ to: "/reviews", search: { problem: id }, replace: true }),
    close: () => void router.navigate({ to: "/reviews", search: {}, replace: true }),
  };
}
