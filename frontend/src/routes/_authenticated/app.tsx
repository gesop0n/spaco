import { createFileRoute } from "@tanstack/react-router";
import { WelcomePage } from "./-components/WelcomePage";

export const Route = createFileRoute("/_authenticated/app")({
  component: WelcomePage,
});
