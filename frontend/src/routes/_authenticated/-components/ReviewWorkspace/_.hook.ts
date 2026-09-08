import { useContext } from "react";
import { ReviewWorkspaceContext } from "./context";

export function useReviewWorkspace() {
  const context = useContext(ReviewWorkspaceContext);
  if (!context) throw new Error("ReviewWorkspaceProvider is required");
  return context;
}
