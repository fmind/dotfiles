terraform {
  # OpenTofu reads this same block; the constraint tracks the tested minor.
  required_version = ">= 1.12"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 8.0"
    }
  }

  # Remote state on GCS — ships commented so a fresh scaffold stays local-green.
  # Create the bucket once, uncomment, then re-run `mise run install`:
  #   gcloud storage buckets create gs://<project>-tofu-state --location=<region> --uniform-bucket-level-access
  #   gcloud storage buckets update gs://<project>-tofu-state --versioning
  # backend "gcs" {
  #   bucket = "<project>-tofu-state"
  #   prefix = "<slug>"
  # }

  # OpenTofu-only: client-side state/plan encryption — state at rest in the
  # bucket becomes ciphertext, so a leaked bucket no longer leaks secrets.
  # Requires the KMS key to exist and every operator to reach it.
  # encryption {
  #   key_provider "gcp_kms" "state" {
  #     kms_encryption_key = "projects/<project>/locations/<region>/keyRings/tofu/cryptoKeys/state"
  #     key_length         = 32
  #   }
  #   method "aes_gcm" "state" {
  #     keys = key_provider.gcp_kms.state
  #   }
  #   state {
  #     method   = method.aes_gcm.state
  #     enforced = true
  #   }
  #   plan {
  #     method   = method.aes_gcm.state
  #     enforced = true
  #   }
  # }
}
