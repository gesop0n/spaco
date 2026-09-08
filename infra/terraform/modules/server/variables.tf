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
  description = "Artifact Registry repository used by Cloud Run source deploys"
  type        = string
}

variable "release_artifact_repository_id" {
  description = "Artifact Registry repository used by the backend release pipeline"
  type        = string
}

variable "deployer_service_account_id" {
  description = "Account ID of the GitHub Actions deployment service account"
  type        = string
}

variable "workload_identity_pool_id" {
  description = "Workload Identity Pool ID used by GitHub Actions"
  type        = string
}

variable "workload_identity_provider_id" {
  description = "GitHub OIDC provider ID in the Workload Identity Pool"
  type        = string
}

variable "github_repository" {
  description = "GitHub repository in owner/name format"
  type        = string

  validation {
    condition     = can(regex("^[^/]+/[^/]+$", var.github_repository))
    error_message = "github_repository must use the owner/name format."
  }
}

variable "github_repository_id" {
  description = "Immutable numeric GitHub repository ID"
  type        = string

  validation {
    condition     = can(regex("^[1-9][0-9]*$", var.github_repository_id))
    error_message = "github_repository_id must be a positive numeric GitHub repository ID."
  }
}

variable "github_ref" {
  description = "Only Git ref allowed to assume the deployment identity"
  type        = string

  validation {
    condition     = startswith(var.github_ref, "refs/heads/")
    error_message = "github_ref must be a fully qualified branch ref."
  }
}

variable "github_environment" {
  description = "Only GitHub environment allowed to assume the deployment identity"
  type        = string
}

variable "runtime_service_account_id" {
  description = "Account ID of the Cloud Run runtime service account"
  type        = string
}

variable "database_secret_id" {
  description = "Secret Manager secret ID containing DATABASE_URL"
  type        = string
}

variable "database_secret_version" {
  description = "Secret Manager version exposed to the Cloud Run service"
  type        = string

  validation {
    condition     = can(regex("^(latest|[1-9][0-9]*)$", var.database_secret_version))
    error_message = "database_secret_version must be latest or a positive integer."
  }
}

variable "bootstrap_image" {
  description = "Existing image used only when Terraform creates the Cloud Run service"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]+-docker\\.pkg\\.dev/[a-z0-9-]+/.+@sha256:[0-9a-f]{64}$", var.bootstrap_image))
    error_message = "bootstrap_image must be an immutable Artifact Registry image digest."
  }
}

variable "cors_allowed_origins" {
  description = "Exact browser origins allowed to call the API"
  type        = list(string)

  validation {
    condition = (
      length(var.cors_allowed_origins) == length(distinct(var.cors_allowed_origins)) &&
      alltrue([for origin in var.cors_allowed_origins : can(regex("^https?://[^/]+$", origin))])
    )
    error_message = "cors_allowed_origins must contain unique origins without paths or trailing slashes."
  }
}

variable "supabase_url" {
  description = "Supabase project URL used to validate access tokens"
  type        = string

  validation {
    condition     = can(regex("^https://[^/]+\\.supabase\\.co$", var.supabase_url))
    error_message = "supabase_url must be an HTTPS Supabase project origin without a trailing slash."
  }
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
