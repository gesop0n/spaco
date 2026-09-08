# spaco server infrastructure

`spaco`のCloud Run APIと依存するGCPリソースをTerraformで管理する。共通構成を`modules/server`、環境固有のroot moduleを`environments/production`へ分けている。

```text
infra/terraform/
├── modules/
│   └── server/
└── environments/
    └── production/
```

## 管理範囲

| 対象 | 管理元 |
| --- | --- |
| Cloud Runの設定、runtime identity、secret参照 | Terraform |
| Artifact Registry、deployer identity、WIF、IAM | Terraform |
| Secret Managerのpayloadとversion | Terraform管理外 |
| container image、Cloud Run revision、traffic | GitHub Actions |

TerraformはCloud Runのimageとtrafficを`ignore_changes`に含めている。アプリのリリース後にTerraformを実行しても、稼働中revisionやtrafficを以前の状態へ戻さない。

## Terraform

repository rootのNix開発環境からproduction root moduleを操作する。

```sh
nix develop
cd infra/terraform/environments/production
terraform fmt -recursive ../..
terraform validate
terraform plan
terraform apply
```

stateは`spaco-prod-terraform-state-237242555322`の`server/production`へ保存する。planを確認し、意図しない変更や削除がない場合だけapplyする。

## Backend release

[`backend-deploy.yml`](../../.github/workflows/backend-deploy.yml)は`main`へのbackend関連変更、または手動実行でproductionへリリースする。

1. Nix環境でtestとvetを実行する。
2. container imageをbuildし、commit SHA tagで`spaco-backend`へpushする。
3. image digestを指定し、`spaco-api-<short-sha>-<run-number>-<attempt>` revisionをtraffic 0%で作成する。
4. `candidate` tagのURLで`/health`を確認する。
5. 成功したrevisionだけへtrafficを100%切り替える。

GitHub ActionsからGoogle CloudへはWorkload Identity Federationで接続する。サービスアカウント鍵やGitHub Secretは不要で、`gesop0n/spaco`のimmutable repository ID、`main` branch、`production` environmentが一致するjobだけがdeployer identityを引き受けられる。

`production` environmentが未作成なら、最初のworkflow実行時にGitHubが作成する。repository adminは必要に応じてrequired reviewerやdeployment branch protectionを設定する。WIF/IAMの作成直後は反映に数分かかることがある。

## Rollback

GitHub Actionsの`Roll back backend` workflowを`main`から実行し、戻したいCloud Run revisionの完全な名前を入力する。workflowはrevisionの存在を確認してから、そのrevisionへtrafficを100%切り替える。imageの再buildやTerraform applyは行わない。

revision名は次のコマンドでも確認できる。

```sh
gcloud run revisions list \
  --project spaco-prod \
  --region asia-northeast1 \
  --service spaco-api
```
