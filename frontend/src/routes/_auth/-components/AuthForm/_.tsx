import { Link } from "@tanstack/react-router";
import { useAuthForm } from "./_.hook";
import type { AuthMode } from "./types";

const inputClassName =
  "min-h-12.5 w-full rounded-lg border border-[#d5deef] bg-surface px-3.5 py-3 text-base leading-6 text-fg-strong outline-none transition-[border-color,box-shadow] duration-160 placeholder:text-fg/80 hover:border-accent-line focus:border-accent focus:ring-3 focus:ring-accent/15 aria-invalid:border-danger aria-invalid:focus:ring-danger/15 motion-reduce:transition-none";

export function AuthForm({ mode, redirectTo }: { mode: AuthMode; redirectTo?: string }) {
  const isRegistration = mode === "register";
  const {
    recoveryDialogRef,
    errors,
    formError,
    isSubmitting,
    notice,
    passwordVisible,
    emailField,
    passwordField,
    handleSubmit,
    hideFeedback,
    togglePasswordVisibility,
    openRecoveryDialog,
    closeRecoveryDialog,
  } = useAuthForm(mode, redirectTo);

  return (
    <>
      <section
        className="rounded-2xl border border-[#d7e0f0] bg-surface px-8 py-9 shadow-card phone:px-6 phone:py-8"
        aria-labelledby="auth-title"
      >
        <div className="mb-8 text-center">
          <h1 id="auth-title" className="text-2xl leading-9 font-[750] tracking-[-0.035em]">
            {isRegistration ? "spacoをはじめる" : "ログイン"}
          </h1>
          <p className="mt-2 text-sm leading-6 text-fg">
            {isRegistration ? "AtCoderの復習を、習慣に。" : "今日の復習を、はじめよう。"}
          </p>
        </div>

        <form
          method="post"
          noValidate
          onSubmit={handleSubmit}
          onChange={hideFeedback}
          aria-labelledby="auth-title"
        >
          <div>
            <label htmlFor="email" className="mb-2 block text-sm leading-6 font-[600]">
              メールアドレス
            </label>
            <input
              {...emailField}
              id="email"
              type="email"
              inputMode="email"
              autoComplete="username"
              autoCapitalize="none"
              spellCheck={false}
              placeholder="you@example.com"
              required
              className={inputClassName}
              aria-invalid={Boolean(errors.email)}
              aria-describedby={errors.email ? "email-error" : undefined}
            />
            {errors.email && (
              <p id="email-error" className="mt-2 text-xs leading-5 text-danger" aria-live="polite">
                {errors.email.message}
              </p>
            )}
          </div>

          <div className="mt-5">
            <label htmlFor="password" className="mb-2 block text-sm leading-6 font-[600]">
              パスワード
            </label>
            <div className="relative">
              <input
                {...passwordField}
                id="password"
                type={passwordVisible ? "text" : "password"}
                autoComplete={isRegistration ? "new-password" : "current-password"}
                minLength={isRegistration ? 8 : undefined}
                required
                className={`${inputClassName} pr-13`}
                aria-invalid={Boolean(errors.password)}
                aria-describedby={
                  [isRegistration && "password-hint", errors.password && "password-error"]
                    .filter(Boolean)
                    .join(" ") || undefined
                }
              />
              <button
                type="button"
                className="absolute top-1/2 right-1 grid size-11 -translate-y-1/2 place-items-center rounded-md text-fg hover:bg-accent-soft hover:text-accent focus-visible:outline-offset-0"
                aria-label={passwordVisible ? "パスワードを隠す" : "パスワードを表示"}
                aria-pressed={passwordVisible}
                aria-controls="password"
                onClick={togglePasswordVisibility}
              >
                <svg
                  className="size-5 stroke-current stroke-[1.6]"
                  viewBox="0 0 24 24"
                  fill="none"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  aria-hidden="true"
                >
                  <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z" />
                  <circle cx="12" cy="12" r="3" />
                  {passwordVisible && <path d="m3 3 18 18" />}
                </svg>
              </button>
            </div>
            {isRegistration ? (
              <p id="password-hint" className="mt-2 text-xs leading-5 text-fg">
                8文字以上で入力してください
              </p>
            ) : (
              <div className="mt-2 text-right">
                <button
                  type="button"
                  className="rounded-sm text-xs leading-5 text-accent underline-offset-4 hover:underline"
                  aria-haspopup="dialog"
                  onClick={openRecoveryDialog}
                >
                  パスワードを忘れた方
                </button>
              </div>
            )}
            {errors.password && (
              <p
                id="password-error"
                className="mt-2 text-xs leading-5 text-danger"
                aria-live="polite"
              >
                {errors.password.message}
              </p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="mt-8 flex min-h-12.5 w-full items-center justify-center rounded-[9px] bg-accent px-4 py-3.5 text-sm leading-5 font-[650] text-surface shadow-action transition-[background,opacity] duration-160 hover:bg-accent-hover disabled:opacity-60 motion-reduce:transition-none"
          >
            {isSubmitting ? "送信中…" : isRegistration ? "アカウントを作成" : "ログイン"}
          </button>

          <div role="status" aria-live="polite" aria-atomic="true">
            {notice && (
              <p className="mt-5 rounded-lg bg-accent-soft px-4 py-3 text-sm leading-6 text-fg-strong">
                {notice}
              </p>
            )}
          </div>
          {formError && (
            <p
              className="mt-5 rounded-lg bg-red-50 px-4 py-3 text-sm leading-6 text-danger"
              role="alert"
            >
              {formError}
            </p>
          )}
        </form>
      </section>

      <p className="mt-7 flex flex-wrap items-center justify-center gap-x-2 gap-y-1 text-center text-[13px] leading-6 text-fg">
        <span>{isRegistration ? "登録済みの方は" : "アカウントをお持ちでない方は"}</span>
        <Link
          to={isRegistration ? "/login" : "/register"}
          className="rounded-sm font-semibold text-accent underline-offset-4 hover:underline"
        >
          {isRegistration ? "ログイン" : "新規登録"}
        </Link>
      </p>

      {!isRegistration && (
        <dialog
          ref={recoveryDialogRef}
          className="m-auto max-h-[calc(100svh-48px)] w-[min(400px,calc(100%-40px))] overflow-y-auto rounded-2xl border border-border bg-surface p-7 text-fg-strong shadow-dialog backdrop:bg-[#1e2b44]/28 backdrop:backdrop-blur-[5px]"
          aria-labelledby="recovery-title"
          aria-describedby="recovery-description"
        >
          <h2 id="recovery-title" className="text-xl leading-8 font-bold">
            パスワードの再設定
          </h2>
          <p id="recovery-description" className="mt-3 text-sm leading-7 text-fg">
            パスワードの再設定は現在準備中です。公開までしばらくお待ちください。
          </p>
          <button
            type="button"
            className="mt-6 min-h-11.5 w-full rounded-lg bg-accent-soft px-4 py-3 text-sm font-[650] text-accent hover:bg-accent-line"
            onClick={closeRecoveryDialog}
          >
            閉じる
          </button>
        </dialog>
      )}
    </>
  );
}
