module "server" {
  source = "../../modules/server"

  project_id                 = var.project_id
  region                     = var.region
  environment                = var.environment
  cloud_run_service_name     = var.cloud_run_service_name
  artifact_repository_id     = var.artifact_repository_id
  runtime_service_account_id = var.runtime_service_account_id
  database_secret_id         = var.database_secret_id
  database_secret_version    = var.database_secret_version
  bootstrap_image            = var.bootstrap_image
  cors_allowed_origins       = var.cors_allowed_origins
  supabase_url               = var.supabase_url
  min_instance_count         = var.min_instance_count
  max_instance_count         = var.max_instance_count
  container_concurrency      = var.container_concurrency
  container_cpu              = var.container_cpu
  container_memory           = var.container_memory
}
