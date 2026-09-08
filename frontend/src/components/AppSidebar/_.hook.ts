import { useQuery } from "@connectrpc/connect-query";
import { useLocation, useRouter } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { AccountService } from "@/__generated__/spaco/account/v1/account_pb";
import { useAuth } from "@/components/AuthProvider";
import { useSidebar } from "@/components/ui/sidebar";

export function useAppSidebar() {
  const { signOut } = useAuth();
  const router = useRouter();
  const pathname = useLocation({ select: (location) => location.pathname });
  const { openMobile, setOpenMobile, isMobile, state, toggleSidebar } = useSidebar();
  const { data } = useQuery(AccountService.method.getCurrentAccount, {});
  const [isSigningOut, setIsSigningOut] = useState(false);
  const [signOutError, setSignOutError] = useState<string>();
  const mobileTriggerRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    setOpenMobile(false);
  }, [pathname, setOpenMobile]);

  const handleSignOut = async () => {
    if (isSigningOut) return;
    setIsSigningOut(true);
    setSignOutError(undefined);
    try {
      await signOut();
      await router.navigate({ to: "/", replace: true });
    } catch {
      setSignOutError("ログアウトに失敗しました。もう一度お試しください。");
    } finally {
      setIsSigningOut(false);
    }
  };

  return {
    account: data?.account,
    pathname,
    isSigningOut,
    signOutError,
    handleSignOut,
    closeMobileSidebar: () => setOpenMobile(false),
    toggleSidebar,
    isMobile,
    openMobile,
    mobileTriggerRef,
    isCollapsed: !isMobile && state === "collapsed",
  };
}
