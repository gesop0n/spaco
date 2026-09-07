# spaco backend

Go・ConnectRPC・PostgreSQL・sqlcで構成するAPI serverです。外部認証にはSupabase Authを使用し、JWT検証後はアプリ内の`UserID`だけを各業務moduleへ渡します。

## セットアップ

リポジトリrootのNix開発環境を使用します。

```sh
nix develop
cp backend/.env.example backend/.env
cd backend
make migrate-up
make dev
```

起動後は`http://localhost:8080/healthz`で稼働確認できます。

## 主なMake target

```sh
make help
make dev
make build
make test
make test-integration
make test-race
make check
make generate
make migrate-up
make migrate-down
make migrate-status
make migrate-create name=add_reviews
```

`generate`はrepository rootの[`buf.gen.yaml`](../buf.gen.yaml)を使用し、Goコードを`backend/generated`、TypeScriptコードを`frontend/src/__generated__`へ生成します。

## 認証付きRPCの流れ

1. CORS middlewareが許可済みfrontend originか確認する。
2. authentication interceptorがBearer tokenを取得する。
3. Supabase JWKSで署名・issuer・audience・有効期限を検証する。
4. account moduleが`issuer + subject`からアプリ内`UserID`を解決する。
5. 初回identityならAccountと対応関係をtransaction内で自動作成する。
6. Account Handlerがcontext内の`UserID`を使ってユースケースを実行する。
7. PostgreSQL adapterがsqlcの生成コードを通してDBへアクセスする。

`shared`にはモジュール固有のHandlerやrepositoryを置かず、現在はモジュール間で同じ意味を持つ`identifier.UserID`だけを置いています。より詳しい依存ルールは[`internal/modules/README.md`](internal/modules/README.md)を参照してください。

## DB migration

Migrationは[`migrations`](migrations)をaccount moduleのテーブル定義として管理し、Gooseで適用します。[`Makefile`](Makefile)が`backend/.env`をincludeし、`DATABASE_URL`をGoose用の環境変数へ割り当てます。

Supabaseのtransaction poolerとnamed prepared statementが衝突しないよう、GooseとAPI serverのpgx query modeには`exec`を指定しています。`.env`の`DATABASE_URL`へquery parameterを追加する必要はありません。

`.env`はMakefileとして読み込まれるため、値を`"`で囲まない形式を使用します。接続文字列のパスワードに`@`、`:`、`/`、`?`、`#`、`%`などのURL予約文字を含む場合は、percent-encodingした値を設定してください。`$`はMakeによる展開を避けるため`$$`と記述します。

`make test-integration`も同じ`DATABASE_URL`を使用します。テスト固有のidentityだけを作成・削除しますが、開発・テスト用DBに対して実行してください。

## sqlc

SQL queryは[`internal/modules/account/internal/adapter/postgres/queries`](internal/modules/account/internal/adapter/postgres/queries)で管理します。`make generate-sql`を実行すると、[`sqlc.yaml`](sqlc.yaml)に従って型安全なGoコードが`internal/modules/account/internal/adapter/postgres/sqlc`へ生成されます。

repositoryは生成されたqueryを呼び出し、transaction制御とdomain objectへの変換を担当します。SQLを変更した場合は生成コードも同じcommitへ含めてください。
