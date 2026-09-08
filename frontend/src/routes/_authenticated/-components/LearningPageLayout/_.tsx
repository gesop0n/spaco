import type { ReactNode } from "react";
import { FlaskConical } from "lucide-react";

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
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 rounded-xl border border-primary/15 bg-accent/60 px-4 py-2.5 text-xs leading-5 text-muted-foreground">
        <span className="inline-flex items-center gap-1.5 font-semibold text-primary">
          <FlaskConical className="size-3.5" aria-hidden="true" />
          UIプレビュー
        </span>
        <span>サンプルデータです。変更は画面間で共有され、再読み込みでリセットされます。</span>
      </div>
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
