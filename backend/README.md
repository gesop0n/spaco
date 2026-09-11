# spaco backend

Go・Echo・ConnectRPC・PostgreSQL・sqlcで構成するAPI serverです。EchoはHTTP routingとmiddleware、ConnectRPCはAPIの通信契約を担当します。外部認証にはSupabase Authを使用し、JWT検証後はアプリ内の`UserID`だけを各業務moduleへ渡します。

## セットアップ

リポジトリrootのNix開発環境を使用します。

```sh
nix develop
cp backend/.env.example backend/.env
cd backend
make migrate-up
make dev
```

起動後は`http://localhost:8080/health`で稼働確認できます。Cloud Runで予約されたpathとの衝突を避けるため、`/healthz`ではなく`/health`を使用します。

## コンテナ

Cloud Run向けのコンテナは、repository rootを起点にbuildします。

```sh
docker build --tag spaco-backend:local ./backend
docker run --rm --publish 8080:8080 --env-file backend/.env spaco-backend:local
curl --fail http://localhost:8080/health
```

multi-stage buildで生成した静的なserver binaryだけを、非rootのdistroless imageで実行します。`.env`、build成果物などはimageへ含めません。Cloud Runでは自動設定される`PORT`をlistenし、`SERVER_ADDRESS`が明示されている場合はそちらを優先します。

初回の疎通確認では、Dockerfileを使ったsource deployも利用できます。`DATABASE_URL`は先にSecret Managerへ登録し、平文の環境変数やrepositoryへ置かないでください。

```sh
gcloud run deploy spaco-api \
  --source backend \
  --region REGION \
  --no-invoker-iam-check \
  --set-env-vars ENV=production,CORS_ALLOWED_ORIGINS=https://FRONTEND_ORIGIN,SUPABASE_URL=https://PROJECT_REF.supabase.co \
  --set-secrets DATABASE_URL=spaco-database-url:VERSION \
  --startup-probe httpGet.path=/health,timeoutSeconds=2,periodSeconds=5,failureThreshold=12
```

デプロイ後は表示されたCloud Run URLへ`/health`を付け、`200 OK`と`ok`が返ることを確認します。この手動手順は初回のcontainer・接続確認用です。継続的なデプロイでは、Artifact Registry、Cloud Run、IAM、Secret ManagerをTerraformで管理し、[`backend-deploy.yml`](../.github/workflows/backend-deploy.yml)がcommit SHAごとのimageをbuildして新しいrevisionをデプロイします。詳しいフローとrollback方法は[`infra/terraform/README.md`](../infra/terraform/README.md)を参照してください。

## 主なMake target

```sh
make help
make dev
make catalog-sync
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
6. 各moduleのHandlerがcontext内の`UserID`を使ってユースケースを実行する。
7. PostgreSQL adapterがsqlcの生成コードを通してDBへアクセスする。

ConnectRPCのlogging interceptorは、全RPCのprocedure名、stream種別、protocol、Connect code、処理時間を構造化ログとして記録します。tokenやrequest/response bodyは記録しません。認証Interceptorより外側に置くことで、認証失敗も同じ形式で観測できます。

`shared`にはモジュール固有のHandlerやrepositoryを置かず、現在はモジュール間で同じ意味を持つ`identifier.UserID`だけを置いています。より詳しい依存ルールは[`internal/modules/README.md`](internal/modules/README.md)を参照してください。

## DB migration

Migrationは[`migrations`](migrations)で各moduleのテーブル定義として管理し、Gooseで適用します。[`Makefile`](Makefile)が`backend/.env`をincludeし、`DATABASE_URL`をGoose用の環境変数へ割り当てます。

`00002_add_account_username.sql`はアプリ内のユーザー名を追加します。API更新前に
`make migrate-up`を実行してください。既存のAtCoder IDやタイムゾーンはそのまま保持し、
ユーザー名は未設定として移行します。ユーザー名が未設定のアカウントは次回利用時に
プロフィール設定へ進みます。ユーザー名とタイムゾーンがあれば初期設定は完了し、
AtCoder IDは任意です。

`00004_create_catalog.sql`はAtCoder Problemsから取り込むコンテストと問題を、`00005_create_reviews.sql`は復習対象（Ankiのcardに相当）と再挑戦の記録（Ankiのrevlogに相当）を追加します。moduleの境界を保つため、`review_items`の`user_id`と`problem_id`には、他moduleのテーブルへの外部キーを張りません。

Supabaseのtransaction poolerとnamed prepared statementが衝突しないよう、GooseとAPI serverのpgx query modeには`exec`を指定しています。`.env`の`DATABASE_URL`へquery parameterを追加する必要はありません。

`.env`はMakefileとして読み込まれるため、値を`"`で囲まない形式を使用します。接続文字列のパスワードに`@`、`:`、`/`、`?`、`#`、`%`などのURL予約文字を含む場合は、percent-encodingした値を設定してください。`$`はMakeによる展開を避けるため`$$`と記述します。

`make test-integration`も同じ`DATABASE_URL`を使用します。テスト専用のidentity、ユーザー、問題IDだけを作成・削除しますが、開発・テスト用DBに対して実行してください。

## sqlc

SQL queryは、各moduleの`internal/adapter/postgres/queries`で管理します。`make generate-sql`を実行すると、[`sqlc.yaml`](sqlc.yaml)に従って型安全なGoコードが同じadapterの`sqlc`ディレクトリへ生成されます。

repositoryは生成されたqueryを呼び出し、transaction制御とdomain objectへの変換を担当します。SQLを変更した場合は生成コードも同じcommitへ含めてください。

## コンテスト・問題情報の同期

問題の登録画面で使うコンテストと問題は、AtCoder Problemsの`contests.json`、`problems.json`、`contest-problem.json`から取り込みます。migrationを適用した後に実行してください。

```sh
make catalog-sync
```

取り込みは追加と更新だけを行い、行は削除しません。問題URLから登録した問題は最小限の情報で作られ、次の同期で問題名などが補われます。AtCoder Problemsの利用ルールに合わせ、3つのファイルは1秒以上の間隔を空けて取得します。本番環境での定期実行は、まだ設定していません。

## 復習日の計算

次回の復習日は、AnkiでFSRSを有効にした場合と同じ規則で計算します（[`scheduler`](internal/modules/review/internal/domain/scheduler)）。

- 記憶モデルは、Ankiが使うfsrs-rs 6.6.2のFSRS-6です。既定パラメータ、目標記憶率90%、最大間隔36500日を使います。
- Hard・Good・Easyの間隔の大小関係とfuzzは、Ankiのスケジューラに合わせています。
- 日単位で復習するため、Ankiの学習ステップと、1日未満の間隔で再出題するshort-term schedulingは使いません。間隔は最短1日です。
- 「ACできなかった」と「解説・ヒントを見てAC」はAgain、自力ACは手応えに応じてHard・Good・Easyとして扱います。
