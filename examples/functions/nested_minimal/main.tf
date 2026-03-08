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
    dev  = { api = { port = 8080 }, web = null, cache = {} }
    prod = { api = { port = 8080 }, web = { port = 443 }, cache = {} }
  }
}

# Returns services with null and empty entries removed, preserving the nested structure
output "minimal_services" {
  value = provider::util::nestedMinimal(local.services)
}
