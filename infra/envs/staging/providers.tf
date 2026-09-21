terraform {
  required_version = ">= 1.9, < 2.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.70"
    }
  }

  # Local state on purpose: this configuration is only ever planned from this repository.
  # The real apply path (remote state, real credentials) is an override file that is never
  # committed — see infra/README.md.
  backend "local" {}
}

# Configured so that init, validate and plan need no AWS account: static placeholder keys,
# every credential/account/metadata probe skipped, no data sources anywhere in the modules.
provider "aws" {
  region     = var.region
  access_key = "PLANONLY0000000000000"
  secret_key = "plan-only-placeholder-secret-key-not-a-credential"

  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true
  skip_region_validation      = true

  default_tags {
    tags = {
      Project     = "fieldops"
      Environment = var.env
      ManagedBy   = "terraform"
    }
  }
}
