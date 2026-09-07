export function Orbit() {
  return (
    <>
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_5%_55%,#ddf3fa,transparent_45%),radial-gradient(ellipse_at_95%_90%,#e7e0fc,transparent_45%)]" />
      <svg
        className="absolute inset-0 size-full"
        viewBox="0 0 1440 900"
        preserveAspectRatio="xMidYMid slice"
        fill="none"
        focusable="false"
      >
        <g transform="rotate(-25 720 450)">
          <ellipse
            cx="720"
            cy="450"
            rx="770"
            ry="185"
            stroke="#a5bbfa"
            vectorEffect="non-scaling-stroke"
          />
          <circle cx="-50" cy="450" r="4" fill="#3865e8" />
          <circle cx="720" cy="265" r="4" fill="#3865e8" />
          <circle cx="1264.5" cy="580.8" r="4" fill="#3865e8" />
        </g>
        <g transform="rotate(55 780 430)">
          <ellipse
            cx="780"
            cy="430"
            rx="690"
            ry="260"
            stroke="#b4b9fa"
            vectorEffect="non-scaling-stroke"
          />
          <circle cx="780" cy="690" r="4" fill="#3865e8" />
          <circle cx="1267.9" cy="246.2" r="4" fill="#3865e8" />
        </g>
        <path
          d="M-80 80C180 350 960 540 1530 205"
          stroke="#b3c9fb"
          vectorEffect="non-scaling-stroke"
        />
        <circle cx="201.1" cy="260.5" r="4" fill="#3865e8" />
      </svg>
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_50%_30%,rgb(243_246_253/95%)_0%,transparent_35%)]" />
    </>
  );
}
