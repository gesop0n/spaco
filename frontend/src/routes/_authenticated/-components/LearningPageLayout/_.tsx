import type { ReactNode } from "react";

export function LearningPageLayout({
  eyebrow,
  title,
  description,
  actions,
  children,
}: {
  eyebrow: string;
  title: string;
  description: string;
  actions?: ReactNode;
  children: ReactNode;
}) {
  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-7 px-5 py-9 md:px-10 md:py-12">
      <header className="flex flex-wrap items-end justify-between gap-5">
        <div>
          <p className="mb-2 text-[11px] font-semibold tracking-[0.18em] text-primary">{eyebrow}</p>
          <h1 className="text-3xl font-semibold tracking-tight">{title}</h1>
          <p className="mt-3 text-sm leading-6 text-muted-foreground">{description}</p>
        </div>
        {actions}
      </header>
      {children}
    </div>
  );
}
