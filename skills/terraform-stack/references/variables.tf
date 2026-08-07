variable "project" {
  description = "GCP project ID that owns every resource in this configuration"
  type        = string
}

variable "region" {
  description = "Default GCP region for regional resources"
  type        = string
  default     = "europe-west1"

  validation {
    condition     = can(regex("^[a-z]+-[a-z]+[0-9]$", var.region))
    error_message = "Region must be a valid GCP region name (e.g. europe-west1)."
  }
}

variable "labels" {
  description = "Common labels stamped on every resource for cost attribution"
  type        = map(string)
  default     = {}
}
