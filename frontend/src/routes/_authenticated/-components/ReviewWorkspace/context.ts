import { createContext } from "react";
import type { Dispatch } from "react";
import type { StudyAction, StudyProblem } from "./types";

export const ReviewWorkspaceContext = createContext<{
  problems: StudyProblem[];
  dispatch: Dispatch<StudyAction>;
  today: string;
  timeZone: string;
} | null>(null);
