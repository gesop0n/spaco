output "cloud_run_service_uri" {
  description = "Canonical Cloud Run service URI"
  value       = module.server.cloud_run_service_uri
}

output "runtime_service_account" {
  description = "Cloud Run runtime service account"
  value       = module.server.runtime_service_account
}

output "artifact_registry_repository" {
  description = "Artifact Registry repository ID used by Cloud Run source deploys"
  value       = module.server.artifact_registry_repository
}

output "release_artifact_registry_repository" {
  description = "Artifact Registry repository ID used by the release pipeline"
  value       = module.server.release_artifact_registry_repository
}

output "deployer_service_account" {
  description = "GitHub Actions deployment service account"
  value       = module.server.deployer_service_account
}

output "workload_identity_provider" {
  description = "Full Workload Identity Provider name for google-github-actions/auth"
  value       = module.server.workload_identity_provider
}
