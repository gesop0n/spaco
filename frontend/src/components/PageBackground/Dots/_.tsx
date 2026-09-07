export function Dots() {
  return (
    <div
      className="absolute inset-0"
      style={{ maskImage: "radial-gradient(ellipse at 50% 45%, transparent 20%, #000 90%)" }}
    >
      <div
        className="absolute inset-0 bg-[linear-gradient(135deg,#a8e2f7,#c1bafa_50%,#aae8f5)] opacity-80"
        style={{
          maskImage: [
            "radial-gradient(circle, #000 0 8px, transparent 8.5px)",
            "radial-gradient(circle, #000 0 18px, transparent 18.5px)",
            "radial-gradient(circle, #000 0 4px, transparent 4.5px)",
          ].join(", "),
          maskSize: "88px 88px, 176px 176px, 88px 88px",
          maskPosition: "0 0, 44px 44px, 44px 44px",
        }}
      />
    </div>
  );
}
