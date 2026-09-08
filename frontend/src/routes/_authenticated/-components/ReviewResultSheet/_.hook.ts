import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useReviewWorkspace } from "../ReviewWorkspace";
import {
  addDays,
  isValidLocalDateTime,
  localDateTime,
  previewIntervals,
} from "../ReviewWorkspace/model";
import type { ReviewEntry, ReviewResult, StudyProblem } from "../ReviewWorkspace/types";

type ResultValues = { result: ReviewResult; performedAt: string; note: string };

export function useReviewResultSheet(problem: StudyProblem) {
  const { dispatch, timeZone } = useReviewWorkspace();
  const [saved, setSaved] = useState<ReviewEntry>();
  // 同じパネルでの二重送信はreducer側でも同一の記録として扱う。
  const [entryId] = useState(() => crypto.randomUUID());
  const form = useForm<ResultValues>({
    mode: "onTouched",
    defaultValues: { performedAt: localDateTime(timeZone), note: "" },
  });
  const selectedResult = useWatch({ control: form.control, name: "result" });
  const submit = form.handleSubmit((values) => {
    if (saved || problem.paused) return;
    const entry = { ...values, note: values.note.trim(), id: entryId };
    dispatch({ type: "record", problemId: problem.id, entry });
    setSaved(entry);
  });
  return {
    ...form,
    submit,
    saved,
    timeZone,
    selectedResult,
    nextDay: saved
      ? addDays(saved.performedAt.slice(0, 10), previewIntervals[saved.result])
      : undefined,
    performedAtField: form.register("performedAt", {
      required: "実施日時を入力してください。",
      validate: (value) => {
        if (!isValidLocalDateTime(value)) return "有効な実施日時を入力してください。";
        if (value > localDateTime(timeZone)) return "未来の日時は指定できません。";
        if (value.slice(0, 10) < problem.registeredOn)
          return "登録日以降の日時を指定してください。";
        return true;
      },
    }),
  };
}
