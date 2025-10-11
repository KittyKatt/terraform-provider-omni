# Copyright (c) HashiCorp, Inc.

# Can also be provided with environment variable OMNI_ENDPOINT
variable "omni_uri" {
  type = string
}

# Can also be provided with environment variable OMNI_SERVICE_ACCOUNT_KEY
variable "omni_service_account_key" {
  type = string
}

provider "omni" {
  endpoint            = var.omni_uri
  service_account_key = var.omni_service_account_key
}
