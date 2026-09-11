import { timestampDate } from "@bufbuild/protobuf/wkt";
import { createConnectQueryKey } from "@connectrpc/connect-query";
import { Difficulty, ReviewResult, ReviewService } from "@/__generated__/spaco/review/v1/review_pb";
import type { ReviewItem, ReviewLog } from "@/__generated__/spaco/review/v1/review_pb";
import { localDateOf, localDateTime, outcomeLabel } from "./model";
import type { DifficultyKey, ProblemSummary, ResultKey } from "./types";

/** 登録や記録の後に、今日の復習・一覧・登録状況をまとめて取り直すためのquery key。 */
export const reviewServiceQueryKey = createConnectQueryKey({
  schema: ReviewService,
  cardinality: undefined,
});

export const reviewResults: Record<ResultKey, ReviewResult> = {
  independent: ReviewResult.INDEPENDENT,
  assisted: ReviewResult.ASSISTED,
  unsolved: ReviewResult.UNSOLVED,
};

export const difficulties: Record<DifficultyKey, Difficulty> = {
  hard: Difficulty.HARD,
  good: Difficulty.GOOD,
  easy: Difficulty.EASY,
};

function keyOf<Key extends string, Value>(
  record: Record<Key, Value>,
  value: Value,
): Key | undefined {
  return (Object.keys(record) as Key[]).find((key) => record[key] === value);
}

export function logLabel(log: Pick<ReviewLog, "result" | "difficulty">): string {
  const result = keyOf(reviewResults, log.result);
  return result
    ? outcomeLabel(result, keyOf(difficulties, log.difficulty))
    : "結果を確認できません";
}

/** 実施日を、アカウントのタイムゾーンでの暦日として返す。 */
export function performedOn(log: ReviewLog, timeZone: string): string {
  return log.performedAt ? localDateOf(timestampDate(log.performedAt), timeZone) : "";
}

/** 実施時刻を、アカウントのタイムゾーンでHH:mm形式にして返す。 */
export function performedTime(log: ReviewLog, timeZone: string): string {
  return log.performedAt
    ? localDateTime(timeZone, timestampDate(log.performedAt)).slice(11, 16)
    : "";
}

/** サーバーは常に問題を返すが、型上は省略できるため空の問題で補う。 */
export function itemProblem(item: ReviewItem): ProblemSummary {
  return item.problem ?? { id: "", contestId: "" };
}
