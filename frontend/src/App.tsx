import { TransportProvider } from "@connectrpc/connect-query";
import { QueryClientProvider } from "@tanstack/react-query";
import { createRouter, RouterProvider } from "@tanstack/react-router";
import { AuthProvider } from "@/components/AuthProvider";
import { transport } from "@/lib/connect";
import { queryClient } from "@/lib/query";
import { routeTree } from "./routeTree.gen";

const router = createRouter({ routeTree, scrollRestoration: true });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <TransportProvider transport={transport}>
          <RouterProvider router={router} />
        </TransportProvider>
      </AuthProvider>
    </QueryClientProvider>
  );
}
