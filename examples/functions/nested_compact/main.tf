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
  services = {
    dev  = { api = { port = 8080 }, web = null }
    prod = { api = { port = 8080 }, web = { port = 443 } }
  }
}

# Returns services with null entries removed, preserving the nested structure
output "compact_services" {
  value = provider::util::nestedCompact(local.services)
}
