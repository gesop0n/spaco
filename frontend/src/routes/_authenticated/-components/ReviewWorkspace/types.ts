/** catalog・reviewのどちらのProblemも受け取れる、表示に必要な問題の情報。 */
export type ProblemSummary = {
  id: string;
  contestId: string;
  problemIndex?: string;
  name?: string;
};

export type ResultKey = "independent" | "assisted" | "unsolved";

/** 自力でACできたときの手応え。 */
export type DifficultyKey = "hard" | "good" | "easy";

export type ProblemReference = { contestId: string; problemId: string };
