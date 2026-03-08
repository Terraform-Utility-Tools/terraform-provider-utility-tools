terraform {
  required_providers {
    util = {
      source = "TheWolfNL/utility-tools"
    }
  }
  required_version = ">= 1.8.0"
}

provider "util" {}

locals {
  config = {
    api     = { port = 8080 }
    web     = null
    metrics = {}
    cache   = []
  }
}

# Returns {api = {port = 8080}} — nulls and empty collections removed
output "minimal_config" {
  value = provider::util::minimal(local.config)
}
