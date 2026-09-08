output "cloud_run_service_uri" {
  description = "Canonical Cloud Run service URI"
  value       = module.server.cloud_run_service_uri
}

output "runtime_service_account" {
  description = "Cloud Run runtime service account"
  value       = module.server.runtime_service_account
}

output "artifact_registry_repository" {
  description = "Artifact Registry repository ID"
  value       = module.server.artifact_registry_repository
}
