# server module

spacoのCloud Run API、Artifact Registry、runtime Service Account、database secretのメタデータとIAM、必要APIをまとめる内部moduleです。

Cloud Runのcontainer imageとsource build metadataはrelease pipelineが管理します。Secretのpayloadとversionもこのmoduleでは管理しません。

このdirectoryで直接Terraformを実行せず、`environments/`以下のroot moduleから呼び出します。
