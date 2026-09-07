import { createFileRoute } from "@tanstack/react-router";
import { ProfileForm } from "./-components/ProfileForm";

export const Route = createFileRoute("/_authenticated/profile")({
  component: ProfileForm,
});
