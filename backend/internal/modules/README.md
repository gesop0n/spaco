# Modules

このディレクトリには、spacoの業務モジュールと、複数の業務機能を横断して
ひとつの処理を完結させるapplication moduleを配置する。

モジュールはテーブル、画面、RPC、DDD の集約ごとには分割しない。業務上の言葉とルールが一貫し、同じ理由で変更される責務をひとつの単位とする。集約は、各モジュールのドメインモデルの中で整合性を守る単位として設計する。

## モジュール

| モジュール | 所有する責務・データ |
| --- | --- |
| `authentication` | access tokenの検証、外部認証identityからアプリ内UserIDを確定する処理、認証済みUserIDの受け渡し |
| `account` | アプリ内ユーザー、認証サービスのユーザーIDとの対応、参照するAtCoder ID、タイムゾーン |
| `catalog` | コンテスト・問題の基本情報、検索、問題URLから登録した情報の補完 |
| `submission` | AtCoderの提出履歴、同期進捗、同期失敗、再試行 |
| `review` | 復習対象、復習結果、履歴、次回予定、一時停止・再開 |

この境界で、AtCoder上の提出結果は `submission`、本人が入力する「自力でACできた」などの復習結果は `review` が所有する。後から提出履歴が同期されても、確定済みの復習結果や次回予定を上書きしない。

`authentication` は認証情報そのものを所有する業務モジュールではなく、外部認証と
`account` を接続するapplication moduleである。パスワード、session、MFAなどの
業務ルールは所有せず、SupabaseやConnectRPCへの依存は内部adapterへ閉じ込める。

## 標準構成

各モジュールは、必要になったパッケージだけを次の形で追加する。

```text
modules/
└── review/
    ├── review.go              # 他モジュールに公開する型・操作の契約
    ├── module.go              # 組み立て済みモジュールの公開窓口
    ├── factory/
    │   └── factory.go         # repository、use case、adapterの生成と接続
    └── internal/
        ├── domain/            # 集約、entity、value object、ドメインルール
        ├── application/       # use caseと、それが必要とするport
        └── adapter/
            ├── rpc/           # ConnectRPCのhandlerと変換処理
            └── postgres/      # repositoryのPostgreSQL実装
```

`domain` はDB、HTTP、ConnectRPC、Protobufの生成型に依存させない。`application` はユースケースを実行し、DBや他モジュールに必要な操作を小さなinterfaceとして定義する。interfaceは原則として利用側に置く。

`adapter/rpc` はProtobufの入出力とアプリケーションの型を変換し、ユースケースを呼び出す。入力検証やエラーからConnectのステータスへの変換はここで行い、復習間隔などの業務ルールは置かない。

`adapter/postgres` は `application` が要求するrepository interfaceを実装する。SQLは同adapter内に置き、sqlcの生成コードを通して実行する。SQL、生成された型、DBライブラリ固有の型はこの配下に閉じ込める。

`factory` はモジュール内部の具体的な実装を組み立てる唯一の場所とする。アプリ全体の起動処理はfactoryを呼び、完成したモジュールだけを受け取る。

初期段階では、ファイルが1つしかない層のために細かいサブパッケージを増やさなくてもよい。コード量や責務が増えたときに、この標準構成に沿って分割する。

## `internal` の境界

`backend/internal/` と各モジュール内の `internal/` は、異なる範囲を保護する。

- `backend/internal/` は、バックエンドの実装をほかのGoモジュールから利用できないようにする。
- `modules/review/internal/` は、reviewの内部実装をほかの業務モジュールや起動処理から利用できないようにする。

例えば `modules/review/factory` は `modules/review/internal/...` をimportできるが、`modules/account` と `internal/app` はimportできない。Goのビルド時の制約によって、repositoryやuse caseの実装を迂回して使うことを防ぐ。

## モジュール間の依存ルール

1. 他モジュールから利用できるのは、対象モジュール直下で公開した契約だけとする。
2. 他モジュールの `internal`、adapter、repository、DBテーブルを直接利用・更新しない。
3. モジュール間で必要な操作は、利用側が必要最小限のinterfaceとして定義し、factoryまたはアプリ全体の組み立て処理で実装を注入する。
4. 依存方向を一方向に保ち、循環importを作らない。双方向の連携が必要になった場合は、アプリケーション層でのオーケストレーションやイベントを検討する。
5. トランザクションは原則としてひとつの集約・モジュール内に収める。複数モジュールのテーブルをひとつのrepositoryから更新しない。
6. ProtobufとConnectRPCの生成コードは通信上の契約として扱い、ドメインモデルやモジュール間通信の型として直接使わない。

`shared` には、複数モジュールが同じ意味で利用する小さく安定した値や契約だけを置く。
ConnectRPCのinterceptorや外部providerの実装は、複数箇所から利用される場合でも
`shared` へ置かず、その処理を所有するmoduleのadapterへ配置する。複数モジュールで
似たコードが見つかっただけでは共通化せず、同じ概念と変更理由を共有しているかを
確認してから切り出す。

## モジュールをまたぐ識別子

`UserID` のように複数のモジュールで受け渡す識別子は、`backend/internal/shared/identifier` に値オブジェクトとして置く。

型を共有しても所有権は共有しない。`identifier.UserID` をどのモジュールからも参照できることと、ユーザーを生成・更新できることは別である。ユーザーの発行、認証サービスのユーザーIDとの対応付け、AtCoder IDやタイムゾーンの設定は `account` だけが行う。他モジュールは受け取った識別子を運ぶだけで、存在確認や属性の解釈をしない。

識別子は文字列の別名ではなく、値を包んだ構造体として定義する。他モジュールが任意の文字列から識別子を組み立てられないようにするためである。

```go
// internal/shared/identifier/user.go

// Package identifier は、モジュールをまたいで受け渡す識別子だけを定義する。
package identifier

// UserID はアプリ内ユーザーの識別子。
// 値の発行と、対応するユーザーの生成・更新は account モジュールだけが行う。
type UserID struct {
	value string
}

func ParseUserID(s string) (UserID, error)

func (id UserID) String() string
```

`identifier` は標準ライブラリ以外に依存させない。`sql.Scanner`、`driver.Valuer`、JSONタグ、Protobufの生成型をここに持ち込まない。DBや通信の型との変換は、各モジュールの `adapter/postgres` と `adapter/rpc` で行う。

認証済みユーザーの受け渡しは、`authentication` moduleが
`identifier.UserID` をcontextに載せ、各モジュールの `adapter/rpc` が
`authentication.UserIDFromContext` で取り出す形にする。トークンのsubjectから
アプリ内ユーザーを解決するのは `account` の責務であり、`authentication` は
interfaceを通じてその結果を受け取る。

## `shared` に置いてよいもの

| 値 | 置き場所 | 理由 |
| --- | --- | --- |
| ユーザーID | `shared/identifier` | アプリが発行する不透明な識別子で、属性も解釈ルールも持たない |
| AtCoder ID | `account` | 提出履歴の参照先として扱い、ログインや本人確認とは分けるという扱いが `account` の業務ルール |
| タイムゾーン | `account` | 同上。必要とする側はinterfaceとして定義し、実装を注入して受け取る |
| 問題ID、コンテストID | `catalog` | AtCoder由来の外部識別子であり、問題URLからの解釈と正規化が `catalog` の知識。`catalog` 直下で公開し、利用側が一方向にimportする |
| 復習結果の3択 | `review` | 復習の業務ルールそのもの |

ユーザーIDと問題IDを対称に扱わない。前者はアプリ内で発行する値で、モジュールごとに解釈が変わらない。後者は外部の書式と正規化のルールを伴い、それは `catalog` が持つ知識である。

`shared` に何かを追加する前に、2つ以上のモジュールが同じ理由で同時に変更するかを確認する。AtCoder IDの書式が変わる理由は `account` にしかないため、この問いで除外できる。値オブジェクトの形をしていることは、`shared` に置く理由にならない。
