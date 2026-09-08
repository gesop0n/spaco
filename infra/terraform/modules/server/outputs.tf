output "cloud_run_service_uri" {
  description = "Canonical Cloud Run service URI"
  value       = google_cloud_run_v2_service.api.uri
}

output "runtime_service_account" {
  description = "Cloud Run runtime service account"
  value       = google_service_account.api_runtime.email
}

output "artifact_registry_repository" {
  description = "Artifact Registry repository ID"
  value       = google_artifact_registry_repository.backend.repository_id
}
