import { Link } from "@tanstack/react-router";
import { useProfileForm } from "./_.hook";

const inputClassName =
  "min-h-12 w-full rounded-lg border border-[#d5deef] bg-surface px-3.5 py-3 text-base outline-none hover:border-accent-line focus:border-accent focus:ring-3 focus:ring-accent/15 aria-invalid:border-danger";

export function ProfileForm() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    serverError,
  } = useProfileForm();

  return (
    <main className="grid min-h-svh place-items-center px-6 py-12">
      <section className="w-full max-w-lg rounded-2xl border border-border bg-surface p-8 shadow-card">
        <h1 className="text-2xl font-bold">プロフィール設定</h1>
        <p className="mt-2 text-sm leading-6 text-fg">
          復習時刻の計算に使用するAtCoder IDとタイムゾーンを設定します。
        </p>

        <form className="mt-8" noValidate onSubmit={handleSubmit}>
          <div>
            <label htmlFor="atcoder-id" className="mb-2 block text-sm font-semibold">
              AtCoder ID
            </label>
            <input
              {...register("atcoderId")}
              id="atcoder-id"
              autoComplete="username"
              className={inputClassName}
              aria-invalid={Boolean(errors.atcoderId)}
              aria-describedby={errors.atcoderId ? "atcoder-id-error" : undefined}
            />
            {errors.atcoderId && (
              <p id="atcoder-id-error" className="mt-2 text-xs text-danger">
                {errors.atcoderId.message}
              </p>
            )}
          </div>

          <div className="mt-5">
            <label htmlFor="time-zone" className="mb-2 block text-sm font-semibold">
              タイムゾーン
            </label>
            <input
              {...register("timeZone")}
              id="time-zone"
              autoComplete="off"
              placeholder="Asia/Tokyo"
              className={inputClassName}
              aria-invalid={Boolean(errors.timeZone)}
              aria-describedby={errors.timeZone ? "time-zone-error" : "time-zone-hint"}
            />
            <p id="time-zone-hint" className="mt-2 text-xs leading-5 text-fg">
              IANA形式で入力します（例: Asia/Tokyo）。
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
              className="min-h-11 flex-1 rounded-lg bg-accent px-5 py-2.5 text-sm font-semibold text-surface hover:bg-accent-hover disabled:opacity-60"
            >
              {isSubmitting ? "保存中…" : "保存する"}
            </button>
            <Link
              to="/app"
              className="inline-flex min-h-11 items-center rounded-lg px-4 py-2.5 text-sm font-semibold text-fg hover:bg-accent-soft"
            >
              キャンセル
            </Link>
          </div>
        </form>
      </section>
    </main>
  );
}
