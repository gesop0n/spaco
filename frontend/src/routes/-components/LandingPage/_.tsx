import { Link } from "@tanstack/react-router";
import { PageBackground } from "../../../components/PageBackground";
import { ReviewIllustration } from "../ReviewIllustration";

export function LandingPage() {
  return (
    <PageBackground variant="aurora" intensity="subtle" className="flex min-h-svh flex-col">
      <a
        className="fixed top-3 left-3 z-10 translate-y-[-160%] rounded-lg bg-accent px-4.5 py-3 text-surface focus:translate-y-0"
        href="#main"
      >
        メインコンテンツへ
      </a>
      <header className="relative z-1 border-b border-border bg-surface">
        <div className="mx-auto flex min-h-22 w-[calc(100%-96px)] max-w-280 items-center justify-between phone:min-h-18 phone:w-[calc(100%-40px)]">
          <Link
            className="py-1 font-brand text-[34px] leading-none font-[750] tracking-[-0.065em] phone:text-[29px]"
            to="/"
            aria-label="spaco トップページ"
          >
            spaco<span className="text-accent">.</span>
          </Link>
          <Link
            className="inline-flex min-h-11 items-center rounded-lg border-0 bg-transparent px-4 py-2.5 text-sm leading-normal font-[650] text-accent transition-[background] duration-160 ease-[ease] hover:bg-accent-soft motion-reduce:transition-none phone:px-2.5 phone:text-[13px]"
            to="/login"
          >
            ログイン
          </Link>
        </div>
      </header>

      <main
        id="main"
        className="relative isolate flex flex-1 items-center justify-center overflow-hidden px-6 pt-16 pb-9.5 after:absolute after:-bottom-22 after:left-1/2 after:-z-1 after:h-35 after:w-[150%] after:-translate-x-1/2 after:rounded-[50%_50%_0_0] after:bg-surface after:content-[''] min-[1600px]:pt-22 min-[1600px]:pb-16 phone:px-5 phone:pt-13 phone:pb-5.5"
        tabIndex={-1}
      >
        <section className="w-full max-w-235 text-center" aria-labelledby="hero-title">
          <div className="animate-enter motion-reduce:animate-none">
            <h1
              id="hero-title"
              className="m-0 text-[clamp(48px,calc(32px+3.6vw),84px)] leading-[1.3] font-[750] tracking-[-0.065em]"
            >
              解き直す<span className="text-accent">。</span>
            </h1>
            <p className="mt-4.5 text-[15px] leading-[1.8] tracking-[0.06em] text-fg phone:mt-4 phone:text-[13px]">
              AtCoderの復習を、習慣に。
            </p>
            <Link
              className="mt-7 inline-flex min-h-13 items-center justify-center gap-7.5 rounded-[10px] border border-transparent bg-accent px-6.25 py-3.5 text-[15px] leading-[1.4] font-[650] text-surface shadow-action transition-[background,translate] duration-160 ease-[ease] hover:-translate-y-0.5 hover:bg-accent-hover active:translate-y-0 motion-reduce:transition-none motion-reduce:hover:translate-y-0 phone:mt-6 phone:min-h-12.5 phone:text-[14px]"
              to="/register"
            >
              はじめる
              <svg
                className="size-5 stroke-current stroke-[1.7]"
                viewBox="0 0 24 24"
                fill="none"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M5 12h14m-6-6 6 6-6 6" />
              </svg>
            </Link>
          </div>
          <ReviewIllustration />
        </section>
      </main>

      <footer className="bg-surface">
        <div className="mx-auto flex min-h-19 w-[calc(100%-96px)] max-w-280 items-center justify-between gap-4 phone:min-h-17 phone:w-[calc(100%-40px)]">
          <span className="font-brand text-[21px] font-[750] tracking-[-0.065em] text-fg">
            spaco
          </span>
        </div>
      </footer>
    </PageBackground>
  );
}
