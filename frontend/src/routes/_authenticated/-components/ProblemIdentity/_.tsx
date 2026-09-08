import type { Problem } from "../ReviewWorkspace/types";

export function ProblemIdentity({ problem }: { problem: Problem }) {
  return (
    <div className="flex min-w-0 items-center gap-3">
      <span className="grid size-10 shrink-0 place-items-center rounded-xl border border-primary/15 bg-accent/60 text-sm font-semibold text-primary">
        {problem.index}
      </span>
      <div className="min-w-0">
        <p className="break-words text-sm font-semibold">{problem.title}</p>
        <p className="mt-1 break-all text-[11px] tracking-wide text-muted-foreground">
          {problem.contestId.toUpperCase()} · {problem.id}
        </p>
      </div>
    </div>
  );
}
