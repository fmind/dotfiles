provider "google" {
  project = var.project
  region  = var.region
}

# Starter resource: replace with real infrastructure. Kept concrete so
# `tofu validate`, `tflint`, and `tofu test` exercise a non-empty module.
resource "google_storage_bucket" "assets" {
  name     = "${var.project}-assets"
  location = var.region
  labels   = var.labels

  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"

  versioning {
    enabled = true
  }
}
