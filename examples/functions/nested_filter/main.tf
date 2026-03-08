terraform {
  required_providers {
    util = {
      source = "Terraform-Utility-Tools/utility-tools"
    }
  }
  required_version = ">= 1.8.0"
}

provider "util" {}

locals {
  services = {
    dev = {
      api = { port = 8080, tls = false }
      web = { port = 80, tls = true }
    }
    prod = {
      api = { port = 8080, tls = true }
      web = { port = 443, tls = true }
    }
  }
}

# Returns all services where tls = true, preserving the nested structure
output "tls_services" {
  value = provider::util::nestedFilter(local.services, { tls = true })
}
# Result:
# {
#   dev  = { web  = { port = 80,  tls = true } }
#   prod = { api  = { port = 8080, tls = true }
#            web  = { port = 443,  tls = true } }
# }
