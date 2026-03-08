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
  features = {
    api     = true
    web     = false
    metrics = null
    cache   = "enabled"
  }
}

# Returns {api = true, web = false, cache = "enabled"} — nulls removed
output "compact_map" {
  value = provider::util::compact(local.features)
}
