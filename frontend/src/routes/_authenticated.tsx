import { createFileRoute, Outlet } from "@tanstack/react-router";
import { AppLayout } from "@/components/AppLayout";
import { WithAuthBoundary } from "./_authenticated/-components/AuthBoundary";

/** 配下のすべてのrouteに、session確認と現在accountの確認を適用するpathless layout。 */
export const Route = createFileRoute("/_authenticated")({
  component: () => (
    <WithAuthBoundary>
      <AppLayout>
        <Outlet />
      </AppLayout>
    </WithAuthBoundary>
  ),
});
