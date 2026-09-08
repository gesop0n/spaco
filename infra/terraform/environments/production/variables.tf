variable "project_id" {
  description = "Google Cloud project ID"
  type        = string
}

variable "region" {
  description = "Region for the server resources"
  type        = string
}

variable "environment" {
  description = "Application environment name"
  type        = string
}

variable "cloud_run_service_name" {
  description = "Cloud Run service name"
  type        = string
}

variable "artifact_repository_id" {
  description = "Artifact Registry repository ID"
  type        = string
}

variable "runtime_service_account_id" {
  description = "Cloud Run runtime service account ID"
  type        = string
}

variable "database_secret_id" {
  description = "Secret Manager secret ID containing DATABASE_URL"
  type        = string
}

variable "database_secret_version" {
  description = "Secret Manager version exposed to Cloud Run"
  type        = string
}

variable "bootstrap_image" {
  description = "Existing immutable image used when creating the Cloud Run service"
  type        = string
}

variable "cors_allowed_origins" {
  description = "Exact browser origins allowed to call the production API"
  type        = list(string)
}

variable "supabase_url" {
  description = "Production Supabase project URL"
  type        = string
}

variable "min_instance_count" {
  description = "Minimum number of Cloud Run instances"
  type        = number
}

variable "max_instance_count" {
  description = "Maximum number of Cloud Run instances"
  type        = number
}

variable "container_concurrency" {
  description = "Maximum concurrent requests handled by one instance"
  type        = number
}

variable "container_cpu" {
  description = "Cloud Run container CPU limit"
  type        = string
}

variable "container_memory" {
  description = "Cloud Run container memory limit"
  type        = string
}
