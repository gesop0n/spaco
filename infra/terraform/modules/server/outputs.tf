output "cloud_run_service_uri" {
  description = "Canonical Cloud Run service URI"
  value       = google_cloud_run_v2_service.api.uri
}

output "runtime_service_account" {
  description = "Cloud Run runtime service account"
  value       = google_service_account.api_runtime.email
}

output "artifact_registry_repository" {
  description = "Artifact Registry repository ID used by Cloud Run source deploys"
  value       = google_artifact_registry_repository.backend.repository_id
}

output "release_artifact_registry_repository" {
  description = "Artifact Registry repository ID used by the release pipeline"
  value       = google_artifact_registry_repository.release.repository_id
}

output "deployer_service_account" {
  description = "GitHub Actions deployment service account"
  value       = google_service_account.api_deployer.email
}

output "workload_identity_provider" {
  description = "Full Workload Identity Provider name for google-github-actions/auth"
  value       = google_iam_workload_identity_pool_provider.github.name
}
