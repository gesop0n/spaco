moved {
  from = google_artifact_registry_repository.backend
  to   = module.server.google_artifact_registry_repository.backend
}

moved {
  from = google_cloud_run_v2_service.api
  to   = module.server.google_cloud_run_v2_service.api
}

moved {
  from = google_secret_manager_secret.database_url
  to   = module.server.google_secret_manager_secret.database_url
}

moved {
  from = google_secret_manager_secret_iam_member.api_database_accessor
  to   = module.server.google_secret_manager_secret_iam_member.api_database_accessor
}

moved {
  from = google_service_account.api_runtime
  to   = module.server.google_service_account.api_runtime
}

moved {
  from = google_project_service.required["artifactregistry.googleapis.com"]
  to   = module.server.google_project_service.required["artifactregistry.googleapis.com"]
}

moved {
  from = google_project_service.required["cloudbuild.googleapis.com"]
  to   = module.server.google_project_service.required["cloudbuild.googleapis.com"]
}

moved {
  from = google_project_service.required["iam.googleapis.com"]
  to   = module.server.google_project_service.required["iam.googleapis.com"]
}

moved {
  from = google_project_service.required["run.googleapis.com"]
  to   = module.server.google_project_service.required["run.googleapis.com"]
}

moved {
  from = google_project_service.required["secretmanager.googleapis.com"]
  to   = module.server.google_project_service.required["secretmanager.googleapis.com"]
}

moved {
  from = google_project_service.required["serviceusage.googleapis.com"]
  to   = module.server.google_project_service.required["serviceusage.googleapis.com"]
}

moved {
  from = google_project_service.required["storage.googleapis.com"]
  to   = google_project_service.storage
}
