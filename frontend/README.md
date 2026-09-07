# spaco frontend

未ログインユーザー向けのランディングページを `/` に表示します。淡い青と白を基調にした、PC・スマートフォン対応のUIです。

「ログイン」は `/login`、「はじめる」は `/register` に遷移します。登録・ログインページは共通のレイアウトを使い、メールアドレスとパスワードの入力チェック、パスワードの表示切替に対応しています。

登録・ログインにはSupabase Authを使用します。認証済みrouteでは、Supabaseのaccess tokenをConnectRPCの`Authorization` headerへ設定し、`GetCurrentAccount`でアプリ内accountも確認します。登録時のパスワードは8文字以上、ログイン時は未入力のみを入力エラーとします。

`/app`と`/profile`は認証済みユーザーだけが表示できるrouteです。アプリ内profileが未設定の場合は`/profile`へ移動します。パスワード再設定はまだ未実装です。

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
- `src/routes/_authenticated/app.tsx`・`profile.tsx`：マイページとprofile設定ルート。
- `src/components/AuthProvider/`：Supabase sessionの初期取得と変更監視。
- `src/lib/connect/`：Connect transport、Bearer token付与、RPCエラー判定。
- `src/lib/query/`：アプリ全体で共有するTanStack QueryClient。
- `src/components/PageBackground/`：差し替え可能な6種類の背景。
- `src/index.css`：Tailwindのテーマ、共通ベーススタイル、アニメーション。

画面のスタイルは各コンポーネントのTailwindユーティリティで記述します。既存の表示幅を保つため、`tablet:`（850px以下）、`compact:`（720px以下）、`phone:`（600px以下）、`narrow:`（359px以下）を共通のレスポンシブバリアントとして定義しています。

[TanStack Routerのファイルベースルーティング](https://tanstack.com/router/latest/docs/installation/manual)を使用します。`src/routeTree.gen.ts` はVite起動時に自動生成されるファイルです。型チェックで参照するためGitの管理対象とし、ルートを変更したら開発サーバーを起動して生成結果も含めます。

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
