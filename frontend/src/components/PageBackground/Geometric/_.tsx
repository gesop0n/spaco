export function Geometric() {
  return (
    <>
      <div className="absolute -top-10 -left-10 grid aspect-square w-[min(36vw,490px)] grid-cols-3 grid-rows-3 gap-3 phone:top-2 phone:-left-30 phone:w-64 phone:gap-2">
        <span className="col-span-2 row-span-2 rounded-br-full bg-linear-to-br from-[#cee7fb] to-[#dce5fc]" />
        <span className="col-start-3 row-start-2 rounded-full border-2 border-[#c6e1fb]" />
        <span className="col-start-1 row-start-3 rounded-[20%] bg-linear-to-br from-[#e1edfc] to-[#d6e6fb]" />
        <span className="col-start-2 row-start-3 rounded-bl-full bg-linear-to-br from-[#cbe5fc] to-[#dceafb]" />
        <span className="col-start-3 row-start-3 rounded-tr-full bg-linear-to-br from-[#e4d9fc] to-[#e9e1fc]" />
      </div>
      <div className="absolute -right-8 -bottom-9 grid aspect-square w-[min(36vw,490px)] grid-cols-3 grid-rows-3 gap-3 phone:-right-31 phone:bottom-3 phone:w-64 phone:gap-2">
        <span className="col-start-1 row-start-1 rounded-t-full rounded-bl-full bg-linear-to-br from-[#d6e3fc] to-[#cfe3fa]" />
        <span className="col-span-2 col-start-2 row-span-2 rounded-tl-full bg-linear-to-br from-[#cbe7fa] to-[#e1effb]" />
        <span className="col-start-1 row-start-2 rounded-tl-full bg-linear-to-br from-[#d1e4fc] to-[#d9eafb]" />
        <span className="col-start-1 row-start-3 rounded-full border-2 border-[#d6c7fa]" />
        <span className="col-start-2 row-start-3 rounded-l-full bg-linear-to-br from-[#dacefb] to-[#e7e0fc]" />
        <span className="col-start-3 row-start-3 rounded-tl-full bg-linear-to-br from-[#cee0fb] to-[#e5edfb]" />
      </div>
    </>
  );
}
