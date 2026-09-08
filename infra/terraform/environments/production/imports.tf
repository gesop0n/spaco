import {
  to = module.server.google_artifact_registry_repository.backend
  id = "projects/spaco-prod/locations/asia-northeast1/repositories/cloud-run-source-deploy"
}

import {
  to = module.server.google_service_account.api_runtime
  id = "projects/spaco-prod/serviceAccounts/spaco-api-runtime@spaco-prod.iam.gserviceaccount.com"
}

import {
  to = module.server.google_secret_manager_secret.database_url
  id = "projects/spaco-prod/secrets/spaco-database-url"
}

import {
  to = module.server.google_secret_manager_secret_iam_member.api_database_accessor
  id = "projects/spaco-prod/secrets/spaco-database-url roles/secretmanager.secretAccessor serviceAccount:spaco-api-runtime@spaco-prod.iam.gserviceaccount.com"
}

import {
  to = module.server.google_cloud_run_v2_service.api
  id = "projects/spaco-prod/locations/asia-northeast1/services/spaco-api"
}

import {
  to = google_storage_bucket.terraform_state
  id = "spaco-prod-terraform-state-237242555322"
}
