export function Grid() {
  return (
    <>
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_0%_100%,#cdf2fa,transparent_55%),radial-gradient(ellipse_at_100%_0%,#e3d9fc,transparent_50%)]" />
      <div
        className="absolute inset-0 [--cell:48px] phone:[--cell:36px]"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgb(163 189 235 / 30%) 1px, transparent 1px), linear-gradient(to bottom, rgb(163 189 235 / 30%) 1px, transparent 1px)",
          backgroundSize: "var(--cell) var(--cell)",
          maskImage: "radial-gradient(ellipse at 50% 45%, transparent 18%, #000 80%)",
        }}
      />
    </>
  );
}
