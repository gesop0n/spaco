import { Link } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { PageBackground } from "../../../../components/PageBackground";

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <PageBackground variant="aurora">
      <main className="grid min-h-svh place-items-center px-6 py-16 phone:px-5 phone:py-10">
        <div className="w-full max-w-100">
          <div className="mb-10 text-center phone:mb-8">
            <Link
              to="/"
              className="inline-block rounded-sm font-brand text-[64px] leading-none font-[750] tracking-[-0.065em] phone:text-[52px]"
              aria-label="spaco トップページ"
            >
              spaco<span className="text-accent">.</span>
            </Link>
          </div>
          {children}
        </div>
      </main>
    </PageBackground>
  );
}
