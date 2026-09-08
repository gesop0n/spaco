resource "google_service_account" "api_runtime" {
  project      = var.project_id
  account_id   = var.runtime_service_account_id
  display_name = "spaco API runtime"
  description  = "Runtime identity for the spaco Cloud Run API"

  depends_on = [google_project_service.required["iam.googleapis.com"]]

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_secret_manager_secret" "database_url" {
  project             = var.project_id
  secret_id           = var.database_secret_id
  deletion_protection = true
  deletion_policy     = "PREVENT"
  labels              = local.common_labels

  replication {
    auto {}
  }

  depends_on = [google_project_service.required["secretmanager.googleapis.com"]]

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_secret_manager_secret_iam_member" "api_database_accessor" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.database_url.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = google_service_account.api_runtime.member
}
