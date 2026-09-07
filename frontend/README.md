# spaco frontend

未ログインユーザー向けのランディングページを `/` に表示します。淡い青と白を基調にした、PC・スマートフォン対応のUIです。

「ログイン」は `/login`、「はじめる」は `/register` に遷移します。登録・ログインページは共通のレイアウトを使い、メールアドレスとパスワードの入力チェック、パスワードの表示切替に対応しています。

認証APIとの接続は未実装です。有効な入力で送信すると準備中の案内を表示し、アカウント作成やログインは行いません。「パスワードを忘れた方」も準備中の案内を表示します。登録時のパスワードはUI案に合わせて8文字以上とし、ログイン時は未入力のみチェックします。認証APIを実装する際にサーバー側のルールと揃えます。

## 開発

リポジトリのルートから実行します。Node.js・pnpmは `flake.nix` の開発環境を使用します。

```sh
nix develop
pnpm --dir frontend install
pnpm --dir frontend dev
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
