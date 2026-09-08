import { Link } from "@tanstack/react-router";
import {
  BookOpen,
  CalendarCheck2,
  ChevronsUpDown,
  House,
  LoaderCircle,
  LogOut,
  PanelLeftClose,
  PanelLeftOpen,
  RotateCcw,
  Search,
  UserRound,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarSeparator,
} from "@/components/ui/sidebar";
import { useAppSidebar } from "./_.hook";

const upcomingItems = [
  { label: "今日の復習", icon: CalendarCheck2 },
  { label: "問題を探す", icon: Search },
  { label: "復習リスト", icon: BookOpen },
];

const menuClassName =
  "h-11 gap-3 rounded-xl px-3 text-sm text-muted-foreground transition-colors data-active:bg-sidebar-accent data-active:text-sidebar-accent-foreground data-active:shadow-[inset_0_0_0_1px_var(--sidebar-border)] motion-reduce:transition-none";

/** 認証済み画面で共通のナビゲーションとアカウント操作。SidebarProvider内で使用する。 */
export function AppSidebar() {
  const {
    account,
    pathname,
    isSigningOut,
    signOutError,
    handleSignOut,
    closeMobileSidebar,
    toggleSidebar,
    isMobile,
    openMobile,
    mobileTriggerRef,
    isCollapsed,
  } = useAppSidebar();
  const displayName = account?.username || "プロフィール未設定";
  const toggleLabel = isMobile
    ? "サイドバーを閉じる"
    : isCollapsed
      ? "サイドバーを開く"
      : "サイドバーを折りたたむ";

  return (
    <>
      {isMobile && (
        <Button
          ref={mobileTriggerRef}
          variant="outline"
          size="icon-lg"
          className="fixed bottom-[calc(1.25rem+env(safe-area-inset-bottom))] left-5 z-40 size-11 rounded-xl bg-sidebar text-sidebar-foreground shadow-card md:hidden"
          onClick={toggleSidebar}
          hidden={openMobile}
          aria-label="サイドバーを開く"
          aria-expanded={openMobile}
        >
          <PanelLeftOpen aria-hidden="true" />
        </Button>
      )}
      <Sidebar
        collapsible="icon"
        className="border-sidebar-border"
        mobileFinalFocus={mobileTriggerRef}
      >
        <SidebarHeader className="gap-5 px-4 pt-7 pb-5 group-data-[collapsible=icon]:px-2">
          <Link
            to="/app"
            onClick={closeMobileSidebar}
            aria-label="spaco Welcome"
            className="flex min-w-0 items-center gap-2.5 rounded-lg group-data-[collapsible=icon]:justify-center"
          >
            <span className="grid size-8 shrink-0 place-items-center rounded-xl bg-primary text-primary-foreground shadow-action">
              <RotateCcw className="size-4.5" aria-hidden="true" />
            </span>
            <span className="font-brand text-[30px] leading-none font-[750] tracking-[-0.065em] group-data-[collapsible=icon]:hidden">
              spaco<span className="text-primary">.</span>
            </span>
          </Link>
        </SidebarHeader>

        <SidebarContent>
          <nav aria-label="メインナビゲーション">
            <SidebarGroup>
              <SidebarGroupContent>
                <SidebarMenu>
                  <SidebarMenuItem>
                    <SidebarMenuButton
                      render={<Link to="/app" onClick={closeMobileSidebar} />}
                      isActive={pathname === "/app"}
                      aria-current={pathname === "/app" ? "page" : undefined}
                      tooltip="Welcome"
                      className={menuClassName}
                    >
                      <House aria-hidden="true" />
                      <span>Welcome</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>

            <SidebarGroup className="pt-4">
              <SidebarGroupLabel className="px-3 text-[11px] tracking-[0.12em]">
                学習スペース
              </SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu className="gap-1">
                  {upcomingItems.map(({ label, icon: Icon }) => (
                    <SidebarMenuItem key={label}>
                      <SidebarMenuButton
                        type="button"
                        disabled
                        aria-label={`${label}（準備中）`}
                        className="h-11 gap-3 rounded-xl px-3 text-muted-foreground disabled:opacity-65"
                      >
                        <Icon aria-hidden="true" />
                        <span>{label}</span>
                        <span className="ml-auto rounded-md border border-sidebar-border px-1.5 py-0.5 text-[10px] group-data-[collapsible=icon]:hidden">
                          準備中
                        </span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </nav>
        </SidebarContent>

        <SidebarFooter className="gap-3 pb-4">
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                type="button"
                tooltip={toggleLabel}
                aria-label={toggleLabel}
                aria-expanded={!isCollapsed}
                className={menuClassName}
                onClick={toggleSidebar}
              >
                {isCollapsed ? (
                  <PanelLeftOpen aria-hidden="true" />
                ) : (
                  <PanelLeftClose aria-hidden="true" />
                )}
                <span>{toggleLabel}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
          <SidebarSeparator className="mx-0" />
          <DropdownMenu>
            <DropdownMenuTrigger
              aria-label={`${displayName}のアカウントメニュー`}
              render={
                <SidebarMenuButton
                  size="lg"
                  isActive={pathname === "/mypage" || pathname === "/profile"}
                  className="min-w-0 gap-3 rounded-xl p-2 data-open:bg-sidebar-accent group-data-[collapsible=icon]:justify-center"
                />
              }
            >
              <span
                className="grid size-8 shrink-0 place-items-center rounded-full border border-sidebar-border bg-aurora text-xs font-semibold text-primary"
                aria-hidden="true"
              >
                {account?.username
                  ? Array.from(account.username).slice(0, 2).join("").toUpperCase()
                  : "S"}
              </span>
              <span className="min-w-0 flex-1 group-data-[collapsible=icon]:hidden">
                <span className="block truncate text-sm font-medium">{displayName}</span>
                <span className="block truncate text-[11px] leading-5 text-muted-foreground">
                  {account?.timeZone || "ユーザー名を設定しましょう"}
                </span>
              </span>
              <ChevronsUpDown
                className="size-3.5 text-muted-foreground group-data-[collapsible=icon]:hidden"
                aria-hidden="true"
              />
            </DropdownMenuTrigger>
            <DropdownMenuContent
              side={isMobile ? "top" : "right"}
              align="end"
              sideOffset={8}
              className="w-60 max-w-[calc(100vw-2rem)] rounded-xl p-1.5"
            >
              <DropdownMenuGroup>
                <DropdownMenuLabel className="px-3 py-2">
                  <span className="block truncate text-sm font-semibold text-foreground">
                    {displayName}
                  </span>
                  <span className="mt-1 block text-xs font-normal">アカウント</span>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  render={<Link to="/mypage" onClick={closeMobileSidebar} />}
                  disabled={isSigningOut}
                  className="min-h-11 cursor-pointer gap-3 rounded-lg px-3"
                >
                  <UserRound aria-hidden="true" />
                  マイページ
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                variant="destructive"
                disabled={isSigningOut}
                closeOnClick={false}
                onClick={handleSignOut}
                className="min-h-11 cursor-pointer gap-3 rounded-lg px-3"
              >
                {isSigningOut ? (
                  <LoaderCircle
                    className="animate-spin motion-reduce:animate-none"
                    aria-hidden="true"
                  />
                ) : (
                  <LogOut aria-hidden="true" />
                )}
                {isSigningOut ? "ログアウト中…" : "ログアウト"}
              </DropdownMenuItem>
              {signOutError && (
                <p
                  className="m-1 rounded-lg bg-destructive/10 p-3 text-xs leading-5 text-destructive"
                  role="alert"
                >
                  {signOutError}
                </p>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarFooter>
      </Sidebar>
    </>
  );
}
