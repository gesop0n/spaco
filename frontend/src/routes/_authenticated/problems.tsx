import { createFileRoute } from "@tanstack/react-router";
import { ProblemRegistrationPage } from "./-components/ProblemRegistrationPage";

export const Route = createFileRoute("/_authenticated/problems")({
  component: ProblemRegistrationPage,
});
