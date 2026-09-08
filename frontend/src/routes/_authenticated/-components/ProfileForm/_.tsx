import { Link } from "@tanstack/react-router";
import { ChevronDown } from "lucide-react";
import { useProfileForm } from "./_.hook";

const inputClassName =
  "min-h-12 w-full rounded-lg border border-input bg-surface px-3.5 py-3 text-base outline-none hover:border-accent-line focus:border-primary focus:ring-3 focus:ring-primary/15 aria-invalid:border-danger";

const badgeClassName =
  "ml-2 inline-flex items-center rounded-md border px-2 py-0.5 align-middle text-[11px] leading-4 font-medium";
const requiredBadgeClassName = `${badgeClassName} border-danger/20 bg-danger/5 text-danger`;

export function ProfileForm() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    serverError,
    setupCompleted,
    timeZoneGroups,
  } = useProfileForm();

  return (
    <main className="grid min-h-svh place-items-center px-6 py-12">
      <section className="w-full max-w-lg rounded-2xl border border-border bg-surface p-8 shadow-card">
        <h1 className="text-2xl font-bold">プロフィール設定</h1>
        <p className="mt-2 text-sm leading-6 text-fg">
          アプリ内で表示するユーザー名とタイムゾーンを設定します。AtCoder IDは後から追加できます。
        </p>

        <form className="mt-8" noValidate onSubmit={handleSubmit}>
          <div>
            <label htmlFor="username" className="mb-2 block text-sm font-semibold">
              ユーザー名
              <span className={requiredBadgeClassName}>必須</span>
            </label>
            <input
              {...register("username")}
              id="username"
              autoComplete="nickname"
              required
              className={inputClassName}
              aria-invalid={Boolean(errors.username)}
              aria-describedby={errors.username ? "username-error" : "username-hint"}
            />
            <p id="username-hint" className="mt-2 text-xs leading-5 text-fg">
              spacoで表示する名前です。1〜40文字で入力してください。
            </p>
            {errors.username && (
              <p id="username-error" className="mt-2 text-xs text-danger">
                {errors.username.message}
              </p>
            )}
          </div>

          <div className="mt-5">
            <label htmlFor="atcoder-id" className="mb-2 block text-sm font-semibold">
              AtCoder ID
              <span className={`${badgeClassName} border-border bg-muted text-muted-foreground`}>
                任意
              </span>
            </label>
            <input
              {...register("atcoderId")}
              id="atcoder-id"
              autoComplete="off"
              autoCapitalize="none"
              spellCheck={false}
              className={inputClassName}
              aria-invalid={Boolean(errors.atcoderId)}
              aria-describedby={errors.atcoderId ? "atcoder-id-error" : "atcoder-id-hint"}
            />
            <p id="atcoder-id-hint" className="mt-2 text-xs leading-5 text-fg">
              提出履歴を参照するIDです。空欄でも利用でき、空欄で保存すると登録を解除します。
            </p>
            {errors.atcoderId && (
              <p id="atcoder-id-error" className="mt-2 text-xs text-danger">
                {errors.atcoderId.message}
              </p>
            )}
          </div>

          <div className="mt-5">
            <label htmlFor="time-zone" className="mb-2 block text-sm font-semibold">
              タイムゾーン
              <span className={requiredBadgeClassName}>必須</span>
            </label>
            <div className="relative">
              <select
                {...register("timeZone")}
                id="time-zone"
                required
                className={`${inputClassName} cursor-pointer appearance-none pr-10`}
                aria-invalid={Boolean(errors.timeZone)}
                aria-describedby={errors.timeZone ? "time-zone-error" : "time-zone-hint"}
              >
                {timeZoneGroups.map((group) => (
                  <optgroup key={group.label} label={group.label}>
                    {group.options.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </optgroup>
                ))}
              </select>
              <ChevronDown
                aria-hidden="true"
                className="pointer-events-none absolute top-1/2 right-3.5 size-4 -translate-y-1/2 text-fg"
              />
            </div>
            <p id="time-zone-hint" className="mt-2 text-xs leading-5 text-fg">
              お住まいの地域を選択してください。初期値は日本（Asia/Tokyo）です。
            </p>
            {errors.timeZone && (
              <p id="time-zone-error" className="mt-2 text-xs text-danger">
                {errors.timeZone.message}
              </p>
            )}
          </div>

          {serverError && (
            <p className="mt-5 rounded-lg bg-red-50 px-4 py-3 text-sm text-danger" role="alert">
              {serverError}
            </p>
          )}

          <div className="mt-8 flex items-center gap-3">
            <button
              type="submit"
              disabled={isSubmitting}
              className="min-h-11 flex-1 rounded-lg bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground hover:bg-accent-hover disabled:opacity-60"
            >
              {isSubmitting ? "保存中…" : "保存する"}
            </button>
            {setupCompleted && (
              <Link
                to="/app"
                className="inline-flex min-h-11 items-center rounded-lg px-4 py-2.5 text-sm font-semibold text-fg hover:bg-accent-soft"
              >
                キャンセル
              </Link>
            )}
          </div>
        </form>
      </section>
    </main>
  );
}
