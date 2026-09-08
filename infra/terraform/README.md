# spaco server infrastructure

`spaco`のCloud Run APIと依存するGCPリソースをTerraformで管理する。共通構成を`modules/server`、環境固有のroot moduleを`environments/production`へ分けている。

```text
infra/terraform/
├── modules/
│   └── server/
└── environments/
    └── production/
```
