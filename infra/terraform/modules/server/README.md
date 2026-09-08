# server module

spacoのCloud Run API、Artifact Registry、runtime/deployer Service Account、GitHub Actions用Workload Identity Federation、database secretのメタデータとIAM、必要APIをまとめる内部moduleです。

Cloud Runのcontainer image、revision、trafficはrelease pipelineが管理します。Secretのpayloadとversionもこのmoduleでは管理しません。

このdirectoryで直接Terraformを実行せず、`environments/`以下のroot moduleから呼び出します。
