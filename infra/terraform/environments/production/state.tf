resource "google_project_service" "storage" {
  project            = var.project_id
  service            = "storage.googleapis.com"
  disable_on_destroy = false
}

resource "google_storage_bucket" "terraform_state" {
  project                     = var.project_id
  name                        = "spaco-prod-terraform-state-237242555322"
  location                    = "ASIA-NORTHEAST1"
  storage_class               = "STANDARD"
  force_destroy               = false
  public_access_prevention    = "enforced"
  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  soft_delete_policy {
    retention_duration_seconds = 604800
  }

  depends_on = [google_project_service.storage]

  lifecycle {
    prevent_destroy = true
  }
}
