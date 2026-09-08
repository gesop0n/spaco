resource "google_artifact_registry_repository" "backend" {
  project       = var.project_id
  location      = var.region
  repository_id = var.artifact_repository_id
  description   = "Cloud Run Source Deployments"
  format        = "DOCKER"
  mode          = "STANDARD_REPOSITORY"
  labels        = local.common_labels

  depends_on = [google_project_service.required["artifactregistry.googleapis.com"]]

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_artifact_registry_repository" "release" {
  project       = var.project_id
  location      = var.region
  repository_id = var.release_artifact_repository_id
  description   = "spaco backend release images"
  format        = "DOCKER"
  mode          = "STANDARD_REPOSITORY"
  labels        = local.common_labels

  depends_on = [google_project_service.required["artifactregistry.googleapis.com"]]

  lifecycle {
    prevent_destroy = true
  }
}
