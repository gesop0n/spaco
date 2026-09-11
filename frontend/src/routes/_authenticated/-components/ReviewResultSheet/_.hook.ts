import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { Code, ConnectError } from "@connectrpc/connect";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import { skipToken, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { Difficulty, ReviewService } from "@/__generated__/spaco/review/v1/review_pb";
import type { ReviewItem, ReviewLog } from "@/__generated__/spaco/review/v1/review_pb";
import { connectErrorMessage } from "@/lib/connect";
import {
  difficulties,
  performedOn,
  reviewResults,
  reviewServiceQueryKey,
} from "../ReviewWorkspace/api";
import {
  daysBetween,
  isValidLocalDateTime,
  localDateTime,
  zonedDateTimeToDate,
} from "../ReviewWorkspace/model";
import type { DifficultyKey, ResultKey } from "../ReviewWorkspace/types";

type ResultValues = {
  result: ResultKey;
  difficulty: DifficultyKey;
  performedAt: string;
  note: string;
};

export function useReviewResultSheet(item: ReviewItem, timeZone: string) {
  const queryClient = useQueryClient();
  const [saved, setSaved] = useState<ReviewLog>();
  const [serverError, setServerError] = useState<string>();
  // 同じパネルからの再送は、サーバーが同じ記録として扱う。
  const [requestId] = useState(() => crypto.randomUUID());
  const recordReview = useMutation(ReviewService.method.recordReview);
  const form = useForm<ResultValues>({
    mode: "onTouched",
    defaultValues: { difficulty: "good", performedAt: localDateTime(timeZone), note: "" },
  });
  const selectedResult = useWatch({ control: form.control, name: "result" });
  const selectedDifficulty = useWatch({ control: form.control, name: "difficulty" });
  const performedAt = useWatch({ control: form.control, name: "performedAt" });
  const lastReviewedOn = item.lastLog ? performedOn(item.lastLog, timeZone) : undefined;

  function validatePerformedAt(value: string): true | string {
    if (!isValidLocalDateTime(value)) return "有効な実施日時を入力してください。";
    if (value > localDateTime(timeZone)) return "未来の日時は指定できません。";
    const day = value.slice(0, 10);
    if (day < item.registeredOn) return "登録日以降の日時を指定してください。";
    if (lastReviewedOn && day < lastReviewedOn) return "前回の記録以降の日時を指定してください。";
    return true;
  }

  const previewable = !saved && validatePerformedAt(performedAt) === true;
  const previewQuery = useQuery(
    ReviewService.method.previewReviewSchedule,
    previewable
      ? {
          reviewItemId: item.id,
          performedAt: timestampFromDate(zonedDateTimeToDate(performedAt, timeZone)),
        }
      : skipToken,
  );
  const preview = selectedResult
    ? previewQuery.data?.previews.find(
        (candidate) =>
          candidate.result === reviewResults[selectedResult] &&
          candidate.difficulty ===
            (selectedResult === "independent"
              ? difficulties[selectedDifficulty]
              : Difficulty.UNSPECIFIED),
      )
    : undefined;

  const submit = form.handleSubmit(async (values) => {
    if (saved) return;
    setServerError(undefined);
    try {
      const response = await recordReview.mutateAsync({
        reviewItemId: item.id,
        requestId,
        result: reviewResults[values.result],
        difficulty:
          values.result === "independent"
            ? difficulties[values.difficulty]
            : Difficulty.UNSPECIFIED,
        performedAt: timestampFromDate(zonedDateTimeToDate(values.performedAt, timeZone)),
        note: values.note.trim(),
      });
      if (!response.log) throw new Error("record review response has no log");
      setSaved(response.log);
      await queryClient.invalidateQueries({ queryKey: reviewServiceQueryKey });
    } catch (error) {
      setServerError(
        ConnectError.from(error).code === Code.FailedPrecondition
          ? "一時停止中のため記録できません。"
          : connectErrorMessage(error),
      );
    }
  });

  return {
    ...form,
    submit,
    saved,
    savedDaysAfterPerformed: saved
      ? daysBetween(performedOn(saved, timeZone), saved.nextDueOn)
      : undefined,
    serverError,
    timeZone,
    selectedResult,
    preview,
    previewLoading: previewable && previewQuery.isPending,
    previewFailed: previewable && previewQuery.isError,
    performedAtField: form.register("performedAt", {
      required: "実施日時を入力してください。",
      validate: validatePerformedAt,
    }),
  };
}
