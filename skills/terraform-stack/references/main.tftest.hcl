# Mock the provider so plan-only tests never configure a real cloud client.
mock_provider "google" {}

variables {
  project = "test-project"
  region  = "europe-west1"
}

run "bucket_is_locked_down" {
  command = plan

  assert {
    condition     = google_storage_bucket.assets.uniform_bucket_level_access == true
    error_message = "Assets bucket must enforce uniform bucket-level access."
  }

  assert {
    condition     = google_storage_bucket.assets.public_access_prevention == "enforced"
    error_message = "Assets bucket must block public access."
  }
}
