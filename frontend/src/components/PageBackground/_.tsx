import type { ReactNode } from "react";
import { Aurora } from "./Aurora";
import { Ripple } from "./Ripple";
import { Geometric } from "./Geometric";
import { Dots } from "./Dots";
import { Orbit } from "./Orbit";
import { Grid } from "./Grid";
import type { PageBackgroundVariant } from "./types";

const patterns = {
  aurora: Aurora,
  ripple: Ripple,
  geometric: Geometric,
  dots: Dots,
  orbit: Orbit,
  grid: Grid,
} satisfies Record<PageBackgroundVariant, () => ReactNode>;

type Props = {
  children: ReactNode;
  variant?: PageBackgroundVariant;
  intensity?: "normal" | "subtle";
  className?: string;
};

export const PageBackground = ({
  children,
  variant = "aurora",
  intensity = "normal",
  className = "",
}: Props) => {
  const Pattern = patterns[variant];

  return (
    <div className={`relative isolate bg-page ${className}`}>
      <div
        className={`pointer-events-none absolute inset-0 -z-1 overflow-hidden select-none ${intensity === "subtle" ? "opacity-30" : ""}`}
        aria-hidden="true"
      >
        <Pattern />
      </div>
      {children}
    </div>
  );
};
