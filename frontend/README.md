# spaco frontend

未ログインユーザー向けのランディングページを `/` に表示します。Auroraの淡いブルー・シアン・ラベンダーを基調にした、PC・スマートフォン対応のUIです。

「ログイン」は `/login`、「はじめる」は `/register` に遷移します。登録・ログインページは共通のレイアウトを使い、メールアドレスとパスワードの入力チェック、パスワードの表示切替に対応しています。

登録・ログインにはSupabase Authを使用します。認証済みrouteでは、Supabaseのaccess tokenをConnectRPCの`Authorization` headerへ設定し、`GetCurrentAccount`でアプリ内accountも確認します。登録時のパスワードは8文字以上、ログイン時は未入力のみを入力エラーとします。

メール確認後に戻る `/` ではsessionの復元を待ち、認証済みなら自動的に `/app` へ進みます。プロフィール未設定の場合は認証済みrouteの共通判定で `/profile` へ移動するため、追加のボタン操作は不要です。通常の未ログインアクセスではランディングページを表示します。SupabaseのSite URLはアプリのURLを設定してください。[PKCEフロー](https://supabase.com/docs/guides/auth/sessions/pkce-flow)を使用しているため、確認メールは登録操作を行ったブラウザで開いてください。

`/app`（Welcome）、`/mypage`（マイページ）、`/profile`（プロフィール設定）は認証済みユーザーだけが表示できるrouteです。ユーザー名が未設定の場合は`/profile`へ移動します。ユーザー名（前後の空白を除いて1〜40文字、重複可）とタイムゾーンを登録すると利用を開始できます。AtCoder IDは任意で、後から追加・変更・解除できます。空欄で保存すると登録解除になります。サイドバーとマイページにはアプリ内のユーザー名を表示します。パスワード再設定はまだ未実装です。

初回プロフィール設定の保存後はWelcomeへ進みます。通常のプロフィール編集へのリンクはマイページだけに置き、編集の保存・キャンセル後はマイページへ戻ります。URLの直接アクセスや再読み込みは引き続き可能です。

プロフィール設定では必須項目を淡い赤、任意項目をグレーの囲みラベルで表示します。タイムゾーンは `Asia/Tokyo` を初期値とするプルダウンで、主な地域とブラウザが対応するIANAタイムゾーンから選択できます。保存済みの設定はブラウザの地域設定にかかわらず維持し、一覧にない別名も選択肢として保持します。

## 開発

リポジトリのルートから実行します。Node.js・pnpmは `flake.nix` の開発環境を使用します。

```sh
nix develop
pnpm --dir frontend install
cp frontend/.env.example frontend/.env.local
pnpm --dir frontend dev
```

`.env.local`には次の値を設定します。

```dotenv
VITE_API_BASE_URL=http://localhost:8080
VITE_SUPABASE_URL=https://your-project.supabase.co
VITE_SUPABASE_PUBLISHABLE_KEY=your-publishable-key
```

## 確認

```sh
pnpm --dir frontend build
pnpm --dir frontend lint
pnpm --dir frontend fmt:check
```

## 構成

- `src/routes/__root.tsx`：共通ルートと404表示。
- `src/routes/index.tsx`：ランディングページのルート。
- `src/routes/-components/`：ランディングページ本体と復習のイラスト。
- `src/routes/_auth.tsx`：登録・ログインの共通レイアウトを使うパスレスルート。
- `src/routes/_auth/register.tsx`・`login.tsx`：登録・ログインのルート。
- `src/routes/_auth/-components/`：認証画面のレイアウト、フォーム、入力チェック。
- `src/routes/_authenticated.tsx`：session・accountを確認する認証済みパスレスルート。
- `src/routes/_authenticated/app.tsx`・`mypage.tsx`・`profile.tsx`：Welcome、マイページ、プロフィール設定ルート。
- `src/components/AuthProvider/`：Supabase sessionの初期取得と変更監視。
- `src/components/AppLayout/`：認証済み画面の共通レイアウト。Aurora背景とSidebarProvider。
- `src/components/AppSidebar/`：shadcn Sidebarを使った共通ナビゲーション、下部の開閉ボタン、アカウントのDropdown Menu。
- `src/lib/connect/`：Connect transport、Bearer token付与、RPCエラー判定。
- `src/lib/query/`：アプリ全体で共有するTanStack QueryClient。
- `src/components/PageBackground/`：差し替え可能な6種類の背景。
- `src/index.css`：Tailwindのテーマ、共通ベーススタイル、アニメーション。

画面のスタイルは各コンポーネントのTailwindユーティリティで記述します。既存の表示幅を保つため、`tablet:`（850px以下）、`compact:`（720px以下）、`phone:`（600px以下）、`narrow:`（359px以下）を共通のレスポンシブバリアントとして定義しています。

[TanStack Routerのファイルベースルーティング](https://tanstack.com/router/latest/docs/installation/manual)を使用します。`src/routeTree.gen.ts` はVite起動時に自動生成されるファイルです。型チェックで参照するためGitの管理対象とし、ルートを変更したら開発サーバーを起動して生成結果も含めます。

## サイドバー

認証済みパスレスルートで `AppLayout` を使用し、`/app`、`/mypage`、`/profile` に同じサイドバーを表示します。開閉操作はサイドバー下部に集約し、PCではアイコンだけに折りたためます。AppLayoutのヘッダーとロゴ横には開閉ボタンを置きません。モバイルでサイドバーを閉じている間は、AppSidebarが画面左下に開くボタンを表示します。リンク先へ移動するとモバイルメニューは閉じます。

ロゴと「Welcome」は仮のホーム画面 `/app` へ移動します。「今日の復習」「問題を探す」「復習リスト」はまだ画面がないため、準備中の無効項目として表示します。画面を追加したら `AppSidebar/_.tsx` の項目をTanStack Routerの `Link` に接続します。

一番下のアカウントボタンを押すと [shadcn Dropdown Menu](https://ui.shadcn.com/docs/components/base/dropdown-menu) が開き、「マイページ」と「ログアウト」を選べます。折りたたみ中もアバターから同じメニューを開けます。ログアウト処理中はメニュー項目を無効化します。セッションを維持したまま処理が失敗した場合はメニュー内で再試行でき、Supabase側でセッションが破棄された場合は認証ガードがログイン画面へ戻します。サイドバーにはプロフィール編集への直接リンクを置きません。

`AppSidebar` を個別に使用する場合は、認証済みの `AuthProvider`・Connect Query環境と `SidebarProvider` の内側に置きます。[shadcn Sidebar](https://ui.shadcn.com/docs/components/base/sidebar)のBase UI版を使用しているため、リンクの合成には `render={<Link ... />}` を使います。

## Aurora カラープリセット

`src/index.css` を配色の定義元にしています。`--aurora-*` は背景のSVGと `bg-aurora` のグラデーションで共有し、操作UIにはshadcnの意味別トークンを使用します。

| 用途                 | Tailwindクラス                                                    |
| -------------------- | ----------------------------------------------------------------- |
| 主要ボタン・リンク   | `bg-primary text-primary-foreground` / `text-primary`             |
| 淡い選択・ホバー背景 | `bg-accent text-accent-foreground`                                |
| ページ・カード       | `bg-background` / `bg-card text-card-foreground`                  |
| 補助テキスト・境界線 | `text-muted-foreground` / `border-border`                         |
| Auroraの装飾背景     | `bg-aurora`                                                       |
| サイドバー           | `bg-sidebar` / `bg-sidebar-accent text-sidebar-accent-foreground` |

既存の `bg-page`・`bg-surface`・`text-fg` なども同じトークンに接続しています。`accent` は淡い面の色なので、白文字の主要ボタンには `primary` を使います。`.dark` 用のトークンもAurora系で定義していますが、テーマ切り替えUIは設けていません。

## 背景の切り替え

登録・ログインの背景は `src/routes/_auth/-components/AuthLayout/_.tsx` の `PageBackground` で指定します。`variant` を変えると両方のページに反映されます。現在は `aurora` を使用しています。

オーロラは曲線状のグラデーションの帯と白い境界で描画しています。ランディングページでも `intensity="subtle"` を指定し、背景の装飾のみを30%の不透明度で表示しています。通常の濃さは `intensity="normal"`（既定値）です。

```tsx
import { PageBackground } from "./components/PageBackground";

<PageBackground variant="aurora">
  <main>{/* ページの内容 */}</main>
</PageBackground>;
```

| variant     | 背景                                       |
| ----------- | ------------------------------------------ |
| `aurora`    | オーロラ：青・シアン・薄紫のグラデーション |
| `ripple`    | 波紋：大きな同心円                         |
| `geometric` | ジオメトリ：角丸の図形                     |
| `dots`      | 水玉：大きさの異なるドット                 |
| `orbit`     | オービット：細い曲線と点                   |
| `grid`      | グリッド：方眼と淡いグラデーション         |

背景はCSSとSVGで描画します。装飾部分のみをクリップし、内容のフォーカス表示やスクロールを妨げない構成です。`PageBackground` 自体は高さを固定しないため、ページ側で `min-h-svh` などを指定します。
