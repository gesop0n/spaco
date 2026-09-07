# フロントエンド設計

React + Vite を使用する。ルーティングには TanStack Router のファイルベースルーティングを使用する。

## ディレクトリ構成

```text
src/
├─ routes/
│  ├─ __root.tsx
│  └─ posts/
│     ├─ route.tsx
│     ├─ index.tsx
│     ├─ $postId.tsx
│     └─ -components/
│        └─ PostList/
│           ├─ index.ts
│           ├─ _.tsx
│           └─ _.hook.ts
└─ components/
   └─ Button/
      ├─ index.ts
      ├─ _.tsx
      └─ _.hook.ts
```

## ルーティング

- ルートファイルは `src/routes` 配下に置く。
- ルート定義、loader、search params など、ルーティングに関する処理はルートファイルに記述する。
- 小さなページはルートファイル内に直接実装してよい。大きくなった場合はルート固有コンポーネントへ切り出す。
- ルート固有コンポーネントは、利用するルートに最も近い `-components` 配下に置く。
- TanStack Router では `-` から始まるファイルとディレクトリがルート生成から除外されるため、`_components` や接頭辞のない `components` は使用しない。`_` はパスレスレイアウト用の記号として扱われる。

## コンポーネント

- コンポーネントごとに `ComponentName/` ディレクトリを作成する。
- `index.ts` を外部公開用のエントリーポイントとし、利用側は原則としてディレクトリから import する。
- `_.tsx` には JSX、スタイル、イベントとロジックの接続など、表示に関する実装を記述する。
- `_.hook.ts` には状態管理、副作用、画面固有のロジックを記述する。分離するロジックがない場合は省略してよい。
- `_.tsx` から `_.hook.ts` を参照するときは、循環参照を避けるため `index.ts` を経由せず `./_.hook` から直接 import する。
- React に依存しない純粋関数、型、API 通信などは、責務に応じた名前の `.ts` ファイルへ分離する。

```ts
// PostList/index.ts
export { PostList } from "./_";
```

```tsx
// 利用側
import { PostList } from "./-components/PostList";
```

## 共有範囲

- `src/components` には、アプリ全体で再利用する汎用 UI コンポーネントを置く。
- 特定のルートだけで使用するコンポーネントは、最初から `src/components` に置かず、そのルートの `-components` に置く。
- 複数の独立したルートで再利用されるようになった時点で、`src/components` への移動を検討する。
- `src/components` から `src/routes/**/-components` を import しない。
