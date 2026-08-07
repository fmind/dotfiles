output "assets_bucket" {
  description = "Name of the starter assets bucket"
  value       = google_storage_bucket.assets.name
}
