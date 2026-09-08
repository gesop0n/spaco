resource "google_service_account" "api_deployer" {
  project      = var.project_id
  account_id   = var.deployer_service_account_id
  display_name = "spaco API deployer"
  description  = "Deployment identity assumed by GitHub Actions"

  depends_on = [google_project_service.required["iam.googleapis.com"]]

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_iam_workload_identity_pool" "github" {
  project                   = var.project_id
  workload_identity_pool_id = var.workload_identity_pool_id
  display_name              = "GitHub Actions"
  description               = "GitHub Actions identities allowed to deploy spaco"

  depends_on = [google_project_service.required["iam.googleapis.com"]]

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_iam_workload_identity_pool_provider" "github" {
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = var.workload_identity_provider_id
  display_name                       = "${var.github_repository} ${var.github_environment}"
  description                        = "GitHub OIDC provider restricted to the production branch and environment"

  attribute_mapping = {
    "google.subject"          = "assertion.sub"
    "attribute.environment"   = "assertion.environment"
    "attribute.ref"           = "assertion.ref"
    "attribute.repository"    = "assertion.repository"
    "attribute.repository_id" = "assertion.repository_id"
  }

  attribute_condition = join(" && ", [
    "assertion.repository_id == '${var.github_repository_id}'",
    "assertion.repository == '${var.github_repository}'",
    "assertion.ref == '${var.github_ref}'",
    "assertion.environment == '${var.github_environment}'",
  ])

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_service_account_iam_member" "github_deployer" {
  service_account_id = google_service_account.api_deployer.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository_id/${var.github_repository_id}"
}

resource "google_artifact_registry_repository_iam_member" "deployer_writer" {
  project    = var.project_id
  location   = google_artifact_registry_repository.release.location
  repository = google_artifact_registry_repository.release.repository_id
  role       = "roles/artifactregistry.writer"
  member     = google_service_account.api_deployer.member
}

resource "google_cloud_run_v2_service_iam_member" "deployer" {
  project  = var.project_id
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.developer"
  member   = google_service_account.api_deployer.member
}

resource "google_service_account_iam_member" "deployer_runtime_user" {
  service_account_id = google_service_account.api_runtime.name
  role               = "roles/iam.serviceAccountUser"
  member             = google_service_account.api_deployer.member
}
