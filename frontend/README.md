# spaco frontend

未ログインユーザー向けのランディングページを `/` に表示します。淡い青と白を基調にした、PC・スマートフォン対応のUIです。

「ログイン」「はじめる」は現在、準備中の案内を開きます。認証サービスとの接続は未実装です。

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
- `src/routes/-components/`：ページ本体と復習のイラスト。
- `src/index.css`：Tailwindのテーマ、共通ベーススタイル、アニメーション。

画面のスタイルは各コンポーネントのTailwindユーティリティで記述します。既存の表示幅を保つため、`tablet:`（850px以下）、`compact:`（720px以下）、`phone:`（600px以下）、`narrow:`（359px以下）を共通のレスポンシブバリアントとして定義しています。

[TanStack Routerのファイルベースルーティング](https://tanstack.com/router/latest/docs/installation/manual)を使用します。`src/routeTree.gen.ts` はVite起動時に自動生成されるファイルです。型チェックで参照するためGitの管理対象とし、ルートを変更したら開発サーバーを起動して生成結果も含めます。
