export type Problem = {
  id: string;
  contestId: string;
  index: string;
  title: string;
};

export type ReviewResult = "independent" | "assisted" | "retry";

export type ReviewEntry = {
  id: string;
  result: ReviewResult;
  performedAt: string;
  note: string;
};

export type StudyProblem = Problem & {
  registeredOn: string;
  dueOn: string;
  paused: boolean;
  registrationNote: string;
  history: ReviewEntry[];
};

export type StudyAction =
  | { type: "register"; problems: Problem[]; today: string; note: string }
  | { type: "record"; problemId: string; entry: ReviewEntry }
  | { type: "togglePause"; problemId: string };
