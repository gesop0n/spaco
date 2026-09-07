import { useRef } from "react";

export function useLandingPage() {
  const availabilityDialogRef = useRef<HTMLDialogElement>(null);

  function openAvailabilityDialog() {
    availabilityDialogRef.current?.showModal();
  }

  function closeAvailabilityDialog() {
    availabilityDialogRef.current?.close();
  }

  return { availabilityDialogRef, openAvailabilityDialog, closeAvailabilityDialog };
}
