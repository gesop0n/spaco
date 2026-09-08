project_id                     = "spaco-prod"
region                         = "asia-northeast1"
environment                    = "production"
cloud_run_service_name         = "spaco-api"
artifact_repository_id         = "cloud-run-source-deploy"
release_artifact_repository_id = "spaco-backend"
runtime_service_account_id     = "spaco-api-runtime"
deployer_service_account_id    = "spaco-api-deployer"
workload_identity_pool_id      = "github-actions"
workload_identity_provider_id  = "github"
github_repository              = "gesop0n/spaco"
github_repository_id           = "1359313137"
github_ref                     = "refs/heads/main"
github_environment             = "production"
database_secret_id             = "spaco-database-url"
database_secret_version        = "1"
bootstrap_image                = "asia-northeast1-docker.pkg.dev/spaco-prod/cloud-run-source-deploy/spaco-api@sha256:e576fcaf726ff6cab7886ad023d2d1d49c158aff63a0f1e3d59ef34b5ce68146"
cors_allowed_origins = [
  "http://localhost:5173",
  "https://spaco-frontend.ishikuro6-2.workers.dev",
]
supabase_url          = "https://iufgfmeafdtlhdxratnq.supabase.co"
min_instance_count    = 0
max_instance_count    = 3
container_concurrency = 40
container_cpu         = "1"
container_memory      = "512Mi"
