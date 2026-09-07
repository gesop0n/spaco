import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";
import { LoginPage } from "./-components/LoginPage";

export const Route = createFileRoute("/_auth/login")({
  validateSearch: z.object({
    redirectTo: z.string().optional(),
  }),
  component: LoginPage,
});
