plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

# GCP-aware rules (invalid machine types, deprecated arguments, IAM typos).
# `tflint --init` downloads the pinned release; bump alongside provider majors.
plugin "google" {
  enabled = true
  version = "0.39.0"
  source  = "github.com/terraform-linters/tflint-ruleset-google"
}
