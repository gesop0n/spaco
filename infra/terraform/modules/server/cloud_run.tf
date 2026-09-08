resource "google_cloud_run_v2_service" "api" {
  project              = var.project_id
  name                 = var.cloud_run_service_name
  location             = var.region
  description          = "spaco ConnectRPC API"
  deletion_protection  = true
  ingress              = "INGRESS_TRAFFIC_ALL"
  invoker_iam_disabled = true
  default_uri_disabled = false
  labels               = local.common_labels

  scaling {
    min_instance_count = var.min_instance_count
    max_instance_count = var.max_instance_count
  }

  template {
    labels                           = local.common_labels
    service_account                  = google_service_account.api_runtime.email
    timeout                          = "30s"
    max_instance_request_concurrency = var.container_concurrency

    containers {
      image = var.bootstrap_image

      ports {
        name           = "http1"
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = var.container_cpu
          memory = var.container_memory
        }
        cpu_idle          = true
        startup_cpu_boost = true
      }

      env {
        name  = "ENV"
        value = var.environment
      }

      env {
        name  = "CORS_ALLOWED_ORIGINS"
        value = join(",", var.cors_allowed_origins)
      }

      env {
        name  = "SUPABASE_URL"
        value = var.supabase_url
      }

      env {
        name  = "SUPABASE_JWT_AUDIENCE"
        value = "authenticated"
      }

      env {
        name  = "SUPABASE_JWT_CLOCK_SKEW"
        value = "30s"
      }

      env {
        name  = "SHUTDOWN_TIMEOUT"
        value = "10s"
      }

      env {
        name = "DATABASE_URL"

        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.database_url.secret_id
            version = var.database_secret_version
          }
        }
      }

      startup_probe {
        failure_threshold = 12
        period_seconds    = 5
        timeout_seconds   = 2

        http_get {
          path = "/health"
          port = 8080
        }
      }
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  depends_on = [
    google_project_service.required["run.googleapis.com"],
    google_secret_manager_secret_iam_member.api_database_accessor,
  ]

  lifecycle {
    prevent_destroy = true
    ignore_changes = [
      build_config,
      client,
      client_version,
      template[0].containers[0].image,
    ]
  }
}
