const cards = [
  {
    name: "save",
    title: "問題を残す",
    icon: "M6 4h12v17l-6-4-6 4V4Z",
    className:
      "top-[101px] left-6 w-[252px] -rotate-5 border-border tablet:left-2.5 tablet:w-55 compact:top-[137px] compact:left-0.5 compact:w-37 narrow:w-33.5",
  },
  {
    name: "review",
    title: "もう一度、解く",
    icon: "M20 8a8 8 0 1 0 0 8M20 3v5h-5",
    className:
      "top-6 left-1/2 z-1 w-[270px] -translate-x-1/2 border-input tablet:w-60 compact:top-2.5 compact:w-[174px] narrow:w-42",
  },
  {
    name: "next",
    title: "次の復習へ",
    icon: "M5 5h14v15H5V5Zm3-2v4m8-4v4M5 10h14",
    className:
      "top-[126px] right-6 w-[252px] rotate-5 border-border tablet:right-2.5 tablet:w-55 compact:top-[162px] compact:right-0.5 compact:w-37 narrow:w-33.5",
  },
] as const;

export function ReviewIllustration() {
  return (
    <figure
      className="relative mx-auto mt-11 h-82.5 w-full max-w-225 animate-enter-late text-left select-none [--enter-offset:15px] motion-reduce:animate-none compact:mt-10.5 compact:h-73.75 compact:max-w-95"
      role="img"
      aria-label="問題を残す、もう一度解く、次の復習へ。復習をくり返す流れのイメージ。"
    >
      <div aria-hidden="true">
        <svg
          className="absolute inset-0 size-full overflow-visible compact:top-6.5 compact:left-[-6%] compact:h-60 compact:w-[112%]"
          viewBox="0 0 900 330"
          fill="none"
        >
          <path
            className="stroke-ring/60 stroke-[1.4]"
            vectorEffect="non-scaling-stroke"
            d="M164 193C115 57 387 18 487 70S820 74 799 221C778 358 442 326 178 248"
          />
          <circle className="fill-primary" cx="317" cy="66" r="4" />
          <circle className="fill-primary" cx="785" cy="139" r="4" />
          <circle className="fill-primary" cx="239" cy="268" r="4" />
        </svg>
        {cards.map((card) => (
          <div
            key={card.name}
            className={`absolute rounded-[15px] border bg-surface p-6 shadow-card tablet:p-5 compact:rounded-[11px] compact:p-3.5 narrow:p-3 ${card.className}`}
          >
            <div className="flex items-center gap-3 text-[13px] leading-[1.4] font-[650] whitespace-nowrap compact:gap-1.75 compact:text-[10px]">
              <span
                className={`grid size-9.5 shrink-0 place-items-center rounded-[10px] compact:size-6.75 compact:rounded-[7px] ${card.name === "review" ? "bg-primary text-primary-foreground" : "bg-accent-soft text-primary"}`}
              >
                <svg
                  className="size-5 stroke-current stroke-[1.5] compact:size-4"
                  viewBox="0 0 24 24"
                  fill="none"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d={card.icon} />
                </svg>
              </span>
              <span>{card.title}</span>
            </div>
            {card.name === "next" ? (
              <div className="mt-6 grid grid-cols-[repeat(7,1fr)] gap-2.25 compact:mt-4.5 compact:gap-1.25">
                {Array.from({ length: 14 }, (_, day) => (
                  <span
                    key={day}
                    className={`grid aspect-square place-items-center rounded-[5px] compact:rounded-[3px] ${day === 10 ? "bg-primary" : day >= 7 && day <= 9 ? "bg-accent-soft" : "bg-muted"}`}
                  >
                    {day === 10 && (
                      <svg
                        className="size-4.25 stroke-primary-foreground stroke-2 compact:size-3"
                        viewBox="0 0 24 24"
                        fill="none"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      >
                        <path d="m6 12 4 4 8-8" />
                      </svg>
                    )}
                  </span>
                ))}
              </div>
            ) : (
              <div className="mt-5.75 grid gap-4 compact:mt-4.5 compact:gap-3">
                {[0, 1].map((row) => (
                  <div key={row} className="flex items-center gap-3 compact:gap-2">
                    <span
                      className={`size-6.75 shrink-0 rounded-[7px] compact:size-5 compact:rounded-[5px] ${card.name === "review" && row === 0 ? "bg-accent" : "bg-muted"}`}
                    />
                    <div
                      className={`grid gap-1.75 compact:gap-1.25 ${row === 0 ? "w-28 compact:w-16.75" : "w-21.5 compact:w-12"}`}
                    >
                      <span className="h-1.5 rounded-lg bg-border compact:h-1.25" />
                      <span className="h-1.25 w-[58%] rounded-lg bg-muted compact:h-1" />
                    </div>
                    {card.name === "review" && row === 0 && (
                      <svg
                        className="ml-auto size-5.5 shrink-0 rounded-full bg-primary stroke-primary-foreground stroke-2 p-0.75 compact:size-4.75"
                        viewBox="0 0 24 24"
                        fill="none"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      >
                        <path d="m6 12 4 4 8-8" />
                      </svg>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </figure>
  );
}
