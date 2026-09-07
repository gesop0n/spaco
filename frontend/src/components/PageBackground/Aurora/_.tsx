import { useId } from "react";

export function Aurora() {
  const id = useId();

  return (
    <>
      <div
        className="absolute inset-0"
        style={{
          backgroundImage: [
            "radial-gradient(ellipse at 95% 0%, #bfd6f2 0%, transparent 40%)",
            "radial-gradient(ellipse at 0% 100%, #e4eafb 0%, transparent 35%)",
            "linear-gradient(135deg, #f0f5ff, #fbfdfd 55%, #f2f5fe)",
          ].join(", "),
        }}
      />
      <svg
        className="absolute inset-0 size-full"
        viewBox="0 0 1600 1000"
        preserveAspectRatio="none"
        focusable="false"
      >
        <defs>
          <linearGradient
            id={`${id}-blue`}
            x1="120"
            y1="0"
            x2="580"
            y2="490"
            gradientUnits="userSpaceOnUse"
          >
            <stop stopColor="#9cacfa" />
            <stop offset="0.45" stopColor="#b5cdf8" />
            <stop offset="0.8" stopColor="#e0effb" />
            <stop offset="1" stopColor="#f8fcfd" />
          </linearGradient>
          <linearGradient
            id={`${id}-cyan`}
            x1="100"
            y1="540"
            x2="1470"
            y2="210"
            gradientUnits="userSpaceOnUse"
          >
            <stop stopColor="#b6e7f3" />
            <stop offset="0.35" stopColor="#e4f4fa" />
            <stop offset="0.65" stopColor="#eaf8fa" />
            <stop offset="1" stopColor="#c4eaf3" />
          </linearGradient>
          <linearGradient
            id={`${id}-pink`}
            x1="90"
            y1="850"
            x2="790"
            y2="560"
            gradientUnits="userSpaceOnUse"
          >
            <stop stopColor="#f4dff5" />
            <stop offset="0.4" stopColor="#e7cef6" />
            <stop offset="0.75" stopColor="#e5ddfa" stopOpacity="0.7" />
            <stop offset="1" stopColor="#f6f9fc" stopOpacity="0" />
          </linearGradient>
          <linearGradient
            id={`${id}-lower-cyan`}
            x1="730"
            y1="930"
            x2="1340"
            y2="550"
            gradientUnits="userSpaceOnUse"
          >
            <stop stopColor="#e8f5fb" stopOpacity="0.2" />
            <stop offset="0.45" stopColor="#ceeafa" />
            <stop offset="1" stopColor="#d6eef9" />
          </linearGradient>
          <linearGradient
            id={`${id}-violet`}
            x1="1080"
            y1="620"
            x2="1580"
            y2="960"
            gradientUnits="userSpaceOnUse"
          >
            <stop stopColor="#ecf3ff" stopOpacity="0.4" />
            <stop offset="0.35" stopColor="#d5cdfb" />
            <stop offset="0.8" stopColor="#c9b8f8" />
            <stop offset="1" stopColor="#e3cef9" />
          </linearGradient>
          <filter
            id={`${id}-soften`}
            x="-10%"
            y="-10%"
            width="120%"
            height="120%"
            colorInterpolationFilters="sRGB"
          >
            <feGaussianBlur stdDeviation="14" />
          </filter>
          <filter
            id={`${id}-highlight`}
            x="-10%"
            y="-20%"
            width="120%"
            height="140%"
            colorInterpolationFilters="sRGB"
          >
            <feGaussianBlur stdDeviation="20" />
          </filter>
        </defs>

        <g filter={`url(#${id}-soften)`}>
          <path
            d="M-80-80H1420C1240 55 870 85 645 280C410 485 225 495-80 245Z"
            fill={`url(#${id}-blue)`}
          />
          <path
            d="M-80 255C115 285 220 500 505 510C805 530 1160 220 1680 125V355C1185 330 925 610 575 640C270 665 100 500-80 555Z"
            fill={`url(#${id}-cyan)`}
          />
          <path
            d="M-80 720C140 550 370 555 610 600C805 635 925 620 1070 570C835 770 540 790 335 855C140 915 20 985-80 975Z"
            fill={`url(#${id}-pink)`}
          />
          <path
            d="M-80 1020C185 1010 420 890 665 815C905 735 1000 750 1190 625C1380 500 1510 390 1680 375V680C1450 720 1200 810 965 955L720 1080H-80Z"
            fill={`url(#${id}-lower-cyan)`}
          />
          <path
            d="M740 1080C900 865 1040 720 1210 600C1385 470 1540 455 1680 400V1080Z"
            fill={`url(#${id}-violet)`}
          />
        </g>

        <g
          fill="none"
          stroke="#fff"
          strokeWidth="32"
          strokeOpacity="0.65"
          filter={`url(#${id}-highlight)`}
        >
          <path d="M-80 255C115 285 220 500 505 510C805 530 1160 220 1680 125" />
          <path d="M-80 965C160 1000 420 875 665 805C905 725 1000 740 1190 615C1380 490 1510 380 1680 365" />
        </g>
      </svg>
    </>
  );
}
