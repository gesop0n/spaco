# authentication module

このmoduleは、外部のaccess tokenを受け取り、認証済みのアプリ内
`identifier.UserID`を確定するまでの処理を所有する。

Supabase上のユーザーとアプリ内ユーザーの対応データは所有しない。その対応付けは
`account` moduleの責務であり、`IUserResolver`を通じて利用する。

## 処理の流れ

```text
Authorization header
        │
        ▼
RPC adapter ── Bearer tokenを抽出
        │
        ▼
application.Service.Authenticate
        │
        ├── ITokenVerifier
        │      └── Supabase JWT・JWKS検証
        │
        └── IUserResolver
               └── account moduleがUserIDを解決
        │
        ▼
contextへidentifier.UserIDを格納
        │
        ▼
各moduleのRPC handler
```

## ディレクトリ構成

```text
authentication/
├── README.md
├── authentication.go       # 他moduleへ公開するIUserResolver
├── context.go              # 認証済みUserIDの受け渡し
├── factory/
│   └── factory.go          # use caseとadapterの組み立て
└── internal/
    ├── application/
    │   ├── authenticate.go # 認証use case
    │   ├── identity.go     # module内部の外部認証identity
    │   └── ports.go        # ITokenVerifier
    └── adapter/
        ├── rpc/
        │   └── interceptor.go
        └── supabase/
            └── token_verifier.go
```

## 依存ルール

- applicationはConnectRPC、JWT library、Supabase SDKへ依存しない。
- Supabase adapterは`ITokenVerifier`を実装する。
- RPC adapterはapplication serviceだけを呼び、JWTを直接解釈しない。
- authenticationからaccountの具体型をimportしない。
- accountは`IUserResolver`を実装し、`internal/app`がfactoryへ注入する。
- 外部認証identityはmodule内部だけで利用し、ほかの業務moduleへ渡さない。
- 業務handlerへ渡す値は`identifier.UserID`だけとする。

## `shared` との境界

`shared`には複数moduleで同じ意味を持つ純粋な値だけを置く。JWT検証や
ConnectRPC interceptorはauthentication固有の実装なので、`shared`には置かない。
