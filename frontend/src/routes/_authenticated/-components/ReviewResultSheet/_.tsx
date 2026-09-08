import {
  BookOpen,
  CalendarDays,
  Check,
  CircleCheck,
  ExternalLink,
  RotateCcw,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { ProblemIdentity } from "../ProblemIdentity";
import {
  formatDay,
  localDateTime,
  previewIntervals,
  problemUrl,
  resultLabels,
} from "../ReviewWorkspace/model";
import type { ReviewResult, StudyProblem } from "../ReviewWorkspace/types";
import { useReviewResultSheet } from "./_.hook";

const outcomes: { value: ReviewResult; description: string; icon: typeof Check }[] = [
  { value: "independent", description: "何も見ずに、考え方から再現できた", icon: CircleCheck },
  { value: "assisted", description: "解説やヒントを参考にして正解した", icon: BookOpen },
  { value: "retry", description: "もう少し考える時間が必要だった", icon: RotateCcw },
];
const inputClassName =
  "min-h-11 w-full min-w-0 rounded-lg border border-input bg-card px-3 py-2.5 text-sm outline-none focus:border-primary focus:ring-3 focus:ring-primary/15 aria-invalid:border-danger";

export function ReviewResultSheet({
  problem,
  onClose,
}: {
  problem: StudyProblem;
  onClose: () => void;
}) {
  const form = useReviewResultSheet(problem);
  const { errors, isSubmitting } = form.formState;
  return (
    <Sheet
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <SheetContent
        side="right"
        showCloseButton={false}
        className="gap-0 overflow-y-auto data-[side=right]:w-full data-[side=right]:sm:max-w-lg"
      >
        <SheetHeader className="border-b border-border px-6 py-6 pr-16 sm:px-8">
          <span className="mb-2 text-[10px] font-semibold tracking-[0.16em] text-primary">
            REVIEW SESSION · UIプレビュー
          </span>
          <SheetTitle className="text-xl">
            {form.saved ? "復習を記録しました" : "復習を記録"}
          </SheetTitle>
          <SheetDescription className="mt-2 text-xs leading-6">
            {form.saved
              ? "次の予定を確認して、また少しずつ。"
              : "AtCoderで取り組んだ後、今回の結果を残しましょう。"}
          </SheetDescription>
        </SheetHeader>
        <SheetClose
          render={<Button variant="ghost" size="icon" />}
          className="absolute top-5 right-5 size-10"
          aria-label="復習パネルを閉じる"
        >
          <X aria-hidden="true" />
        </SheetClose>
        <div className="border-b border-border bg-background/70 px-6 py-5 sm:px-8">
          <ProblemIdentity problem={problem} />
        </div>

        {form.saved && form.nextDay ? (
          <div className="flex flex-1 flex-col items-center px-6 py-10 text-center sm:px-8">
            <span className="grid size-16 place-items-center rounded-full bg-accent text-primary">
              <Check className="size-7" aria-hidden="true" />
            </span>
            <h2 className="mt-6 text-xl font-semibold">ひとつ、積み重ねました。</h2>
            <p className="mt-3 text-sm text-muted-foreground">{resultLabels[form.saved.result]}</p>
            <div className="mt-8 w-full rounded-2xl border border-primary/15 bg-aurora p-6">
              <p className="flex items-center justify-center gap-2 text-xs text-muted-foreground">
                <CalendarDays className="size-4" aria-hidden="true" />
                次回の復習予定
              </p>
              <p className="mt-3 text-2xl font-semibold text-primary">{formatDay(form.nextDay)}</p>
              <p className="mt-2 text-xs text-muted-foreground">
                実施日から{previewIntervals[form.saved.result]}日後 · プレビュー用の仮日程
              </p>
            </div>
            {form.saved.note && (
              <p className="mt-6 w-full whitespace-pre-wrap break-words rounded-xl bg-muted/60 p-4 text-left text-sm leading-6 text-muted-foreground">
                {form.saved.note}
              </p>
            )}
            <Button onClick={onClose} className="mt-8 h-11 w-full">
              今日の復習に戻る
            </Button>
            <p className="mt-4 text-xs leading-5 text-muted-foreground">
              記録はこのプレビュー内のみで保持されます。
            </p>
          </div>
        ) : (
          <form
            onSubmit={form.submit}
            noValidate
            className="flex flex-1 flex-col px-6 py-6 sm:px-8"
          >
            <a
              href={problemUrl(problem)}
              target="_blank"
              rel="noreferrer"
              className="inline-flex min-h-11 items-center justify-center gap-2 rounded-lg border border-primary/20 bg-accent/60 px-4 py-3 text-sm font-semibold text-primary hover:bg-accent"
            >
              AtCoderで問題を開く
              <ExternalLink className="size-4" aria-hidden="true" />
            </a>
            <p className="mt-2 text-center text-[11px] leading-5 text-muted-foreground">
              問題を読む・コードを書く・提出する操作はAtCoderで行います。
            </p>
            <fieldset
              className="mt-7"
              aria-describedby={errors.result ? "review-result-error" : undefined}
            >
              <legend className="mb-3 text-sm font-semibold">今回の結果</legend>
              <div className="grid gap-2.5">
                {outcomes.map(({ value, description, icon: Icon }) => (
                  <label key={value} className="cursor-pointer">
                    <input
                      type="radio"
                      value={value}
                      {...form.register("result", { required: "今回の結果を選択してください。" })}
                      className="peer sr-only"
                    />
                    <span className="flex items-center gap-3 rounded-xl border border-border p-4 transition-colors peer-checked:border-primary/50 peer-checked:bg-accent/60 peer-focus-visible:ring-3 peer-focus-visible:ring-primary/25 hover:bg-background">
                      <Icon className="size-5 shrink-0 text-primary" aria-hidden="true" />
                      <span className="flex-1">
                        <span className="block text-sm font-semibold">{resultLabels[value]}</span>
                        <span className="mt-1 block text-xs leading-5 text-muted-foreground">
                          {description}
                        </span>
                      </span>
                      <span
                        className={`grid size-4 shrink-0 place-items-center rounded-full border ${form.selectedResult === value ? "border-primary bg-primary text-primary-foreground" : "border-input"}`}
                      >
                        {form.selectedResult === value && (
                          <Check className="size-2.5" aria-hidden="true" />
                        )}
                      </span>
                    </span>
                  </label>
                ))}
              </div>
              {errors.result && (
                <p id="review-result-error" role="alert" className="mt-2 text-xs text-danger">
                  {errors.result.message}
                </p>
              )}
            </fieldset>
            <label htmlFor="review-time" className="mt-6 mb-2 text-sm font-semibold">
              実施日時{" "}
              <span className="ml-1 text-[11px] font-normal text-muted-foreground">
                {form.timeZone}
              </span>
            </label>
            <input
              id="review-time"
              type="datetime-local"
              {...form.performedAtField}
              max={localDateTime(form.timeZone)}
              className={inputClassName}
              aria-invalid={!!errors.performedAt}
              aria-describedby={errors.performedAt ? "review-time-error" : undefined}
            />
            {errors.performedAt && (
              <p id="review-time-error" role="alert" className="mt-2 text-xs text-danger">
                {errors.performedAt.message}
              </p>
            )}
            <label htmlFor="review-note" className="mt-5 mb-2 text-sm font-semibold">
              振り返りメモ{" "}
              <span className="ml-1 text-xs font-normal text-muted-foreground">任意</span>
            </label>
            <textarea
              id="review-note"
              {...form.register("note", {
                maxLength: { value: 1000, message: "メモは1000文字以内で入力してください。" },
              })}
              rows={3}
              maxLength={1000}
              placeholder="気づいたこと、次に意識したいこと。"
              className={`${inputClassName} resize-y`}
              aria-invalid={!!errors.note}
            />
            {errors.note && (
              <p role="alert" className="mt-2 text-xs text-danger">
                {errors.note.message}
              </p>
            )}
            <div className="mt-6 rounded-xl bg-background p-3 text-xs leading-6 text-muted-foreground">
              次回の予定：
              {form.selectedResult
                ? `実施日から${previewIntervals[form.selectedResult]}日後`
                : "結果を選ぶと表示されます"}
              <br />
              間隔はUI確認用の仮ルールです。提出結果の同期は行いません。
            </div>
            <Button type="submit" disabled={isSubmitting} className="mt-5 h-12 w-full gap-2">
              <Check aria-hidden="true" />
              {isSubmitting ? "記録中…" : "結果を記録する"}
            </Button>
          </form>
        )}
      </SheetContent>
    </Sheet>
  );
}
