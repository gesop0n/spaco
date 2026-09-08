import { createFileRoute } from "@tanstack/react-router";
import { ReviewListPage } from "./-components/ReviewListPage";

export const Route = createFileRoute("/_authenticated/review-list")({ component: ReviewListPage });
