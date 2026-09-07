export function Ripple() {
  return (
    <>
      <div
        className="absolute bottom-0 left-0 aspect-square w-[clamp(580px,90vw,1300px)] -translate-x-[44%] translate-y-[44%] rounded-full"
        style={{
          backgroundImage:
            "repeating-radial-gradient(circle, rgb(118 204 239 / 30%) 0% 8%, rgb(161 217 246 / 13%) 8% 16%)",
          maskImage: "radial-gradient(circle, #000 25%, rgb(0 0 0 / 80%) 50%, transparent 72%)",
        }}
      />
      <div
        className="absolute top-0 right-0 aspect-square w-[clamp(520px,80vw,1150px)] translate-x-[42%] -translate-y-[42%] rounded-full"
        style={{
          backgroundImage:
            "repeating-radial-gradient(circle, rgb(171 148 239 / 25%) 0% 8%, rgb(190 178 250 / 10%) 8% 16%)",
          maskImage: "radial-gradient(circle, #000 25%, rgb(0 0 0 / 80%) 50%, transparent 72%)",
        }}
      />
    </>
  );
}
