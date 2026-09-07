import { useRef } from "react";

export function useLandingPage() {
  const availabilityDialogRef = useRef<HTMLDialogElement>(null);

  const openAvailabilityDialog = () => {
    availabilityDialogRef.current?.showModal();
  };

  const closeAvailabilityDialog = () => {
    availabilityDialogRef.current?.close();
  };

  return { availabilityDialogRef, openAvailabilityDialog, closeAvailabilityDialog };
}
