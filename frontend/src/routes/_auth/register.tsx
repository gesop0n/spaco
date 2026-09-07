import { createFileRoute } from "@tanstack/react-router";
import { AuthForm } from "./-components/AuthForm";

export const Route = createFileRoute("/_auth/register")({
  component: () => <AuthForm key="register" mode="register" />,
});
