import { createFileRoute } from "@tanstack/react-router";
import { ReviewDashboardPage } from "./-components/ReviewDashboardPage";

export const Route = createFileRoute("/_authenticated/reviews")({
  validateSearch: (search: Record<string, unknown>): { problem?: string } => ({
    problem: typeof search.problem === "string" ? search.problem : undefined,
  }),
  component: ReviewDashboardPage,
});
