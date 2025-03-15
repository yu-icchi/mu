terraform {
  backend "local" {}

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.54.1"
    }
  }
}

variable "endpoint" {
  type = string
}

provider "aws" {
  region = "us-east-1"

  access_key = "mock_access_key"
  secret_key = "mock_secret_key"

  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    iam     = var.endpoint
    kinesis = var.endpoint
    lambda  = var.endpoint
    s3      = var.endpoint
  }
}
