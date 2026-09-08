import type { ReactNode } from "react";
import { AppSidebar } from "@/components/AppSidebar";
import { PageBackground } from "@/components/PageBackground";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";

export function AppLayout({ children }: { children: ReactNode }) {
  return (
    <PageBackground className="min-h-svh" intensity="subtle">
      <TooltipProvider>
        <SidebarProvider>
          <a
            href="#app-content"
            className="fixed top-3 left-3 z-50 -translate-y-24 rounded-lg bg-primary px-4 py-3 text-sm text-primary-foreground focus:translate-y-0"
          >
            メインコンテンツへ
          </a>
          <AppSidebar />
          <SidebarInset className="min-w-0 bg-transparent">
            <div
              id="app-content"
              tabIndex={-1}
              className="flex flex-1 flex-col pb-20 outline-none md:pb-0"
            >
              {children}
            </div>
          </SidebarInset>
        </SidebarProvider>
      </TooltipProvider>
    </PageBackground>
  );
}
