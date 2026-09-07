import { createFileRoute } from "@tanstack/react-router";
import { AuthForm } from "./-components/AuthForm";

export const Route = createFileRoute("/_auth/login")({
  component: () => <AuthForm key="login" mode="login" />,
});
