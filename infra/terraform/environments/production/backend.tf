terraform {
  backend "gcs" {
    bucket = "spaco-prod-terraform-state-237242555322"
    prefix = "server/production"
  }
}
