# account モジュール内部の設計

このディレクトリには、`account` モジュールの外部へ公開しない実装を置く。
ほかの業務モジュールや `internal/app` は、この配下を直接 import せず、
`account` 直下で公開された契約を利用する。

## account が所有する責務

`account` は、アプリ内のユーザーと外部認証サービス上のユーザーを対応付ける。

- アプリ内 `UserID` の発行とアカウントの生成
- 認証サービスの `issuer + subject` と `UserID` の対応付け
- 参照対象となる AtCoder ID の設定と変更
- タイムゾーンの設定と変更

`subject` は認証サービス内でのみ一意な場合があるため、外部ユーザーの識別には
`subject` 単体ではなく `issuer + subject` の組を使用する。

## 認証との境界

JWTやCookieの検証は技術的な認証処理であり、`account` のドメインルールではない。
そのため、外部認証との接続と認証処理のオーケストレーションは
`authentication` moduleが担当する。

認証処理は次のように分担する。

1. `authentication` のRPC adapterがAuthorization headerからトークンを取得する。
2. `authentication` のSupabase adapterが署名、issuer、audience、有効期限を検証する。
3. `authentication` use caseが検証済みの `issuer + subject` を
   `account` のuse caseへ渡す。
4. `account` が対応する `identifier.UserID` を解決する。
5. `authentication` のRPC adapterが `UserID` をcontextに載せる。

トークン、JWT claims、認証サービスの SDK 型は domain 層へ持ち込まない。
domain 層が受け取るのは、検証済みの値から生成した `AuthIdentity` とする。

現在の `authentication` はパスワード、session、refresh token、MFA、
アカウントロックなどを所有しない。それらをアプリ自身が所有するようになった場合は、
`authentication` の責務を見直し、`identity` や `access` といった業務モジュールへの
分離を検討する。

## 現在の構成

```text
internal/
├── README.md
├── domain/
│   ├── account.go
│   └── auth_identity.go
├── usecase/
│   ├── ports.go
│   ├── resolve_user.go
│   ├── get_current_account.go
│   └── update_profile.go
└── adapter/
    ├── rpc/handler.go
    └── postgres/
        ├── repository.go      # transaction制御とdomain型への変換
        ├── queries/           # sqlcへ渡すSQL query
        └── sqlc/              # sqlcが生成する型安全なDBアクセスコード
```

## ドメインモデル

最初は `Account` を集約ルートとする。

```go
type Account struct {
	id        identifier.UserID
	atCoderID *string
	timeZone  string
}
```

各値には次の役割を持たせる。

### `identifier.UserID`

アプリ内ユーザーの不透明な識別子。ほかのモジュールとの受け渡しには
`internal/shared/identifier.UserID` を使うが、その発行と対応するアカウントの生成は
`account` だけが行う。

### `AuthIdentity`

外部認証サービスの `issuer + subject` を保持する値オブジェクト。
空文字を許可せず、アカウント作成後は原則として変更しない。

複数の認証方法を同じアカウントへリンクする要件が生まれるまでは、複数件を扱うための
抽象化を先に追加しない。

### `AtCoderID`

提出履歴を参照する対象を表す値オブジェクト。空文字や文字数など、確認できた
AtCoderの制約だけを生成時に検証する。外部APIへの問い合わせが必要な存在確認は
値オブジェクトでは行わず、use caseからportを介して実行する。

AtCoder ID を一意にするかどうかは、本人確認済みアカウントだけを扱うのか、単に
参照したいIDを登録できるのかというプロダクト要件を決めてから判断する。

### `TimeZone`

ユーザーの日付境界や表示時刻に利用するタイムゾーン。`Asia/Tokyo` のような
IANA Time Zone Database の名前として生成時に検証する。

## use case

### `ResolveUser`

検証済みの`AuthIdentity`に対応する`identifier.UserID`を返す。対応がない場合は、
未設定Accountとidentityの対応を同じtransaction内で自動作成する。同時requestは
`auth_identities`の複合主キーで直列化し、重複して作ったAccountを残さない。

### `GetCurrentAccount`

認証interceptorがcontextへ格納した`UserID`からAccountを取得する。RPC adapterは
domain型をProtobuf型へ変換し、repositoryやdomainを通信契約へ依存させない。

### `UpdateProfile`

AtCoder IDとIANA形式のタイムゾーンをdomain層で検証してから保存する。更新後のAccountを
返し、client側は`GetCurrentAccount`のQuery cacheを再取得する。

use caseが必要とするrepository interfaceは利用側で定義し、具体的なSQLや
DBライブラリの型は `adapter/postgres` に閉じ込める。

PostgreSQL adapterではSQLを`queries`へ定義し、sqlcが生成した`sqlc.Queries`を
repositoryから呼び出す。生成コードへdomainルールやtransaction境界を持たせず、
repositoryが複数queryをまとめ、DBの型をdomain型へ変換する。

## domain 層に置かないもの

- JWT、Bearer token、HTTP Cookie
- Auth0、Firebase、CognitoなどのSDK型
- ConnectRPC、Protobufの生成型
- SQL、トランザクション、DBライブラリ固有の型
- contextへの認証情報の格納処理
- 外部APIを使ったAtCoder IDの存在確認

## module外への公開

module外へ公開するのは`account.Module`だけである。`internal/app`はこのModuleから
Connect Handlerを受け取り、認証interceptorを設定する。また、Module自身が
`authentication.IUserResolver`を満たすため、authenticationはaccountの具象型や
PostgreSQL adapterを参照しない。
