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
  flat = {
    "dev/api/port"  = 8080
    "dev/web/port"  = 443
    "prod/api/port" = 8080
    "prod/web/port" = 443
  }
}

# Expands the flat map back into a nested object
output "expanded" {
  value = provider::util::expand(local.flat)
}
# Result:
# {
#   dev  = { api = { port = 8080 }, web = { port = 443 } }
#   prod = { api = { port = 8080 }, web = { port = 443 } }
# }

# Custom separator
output "expanded_dot" {
  value = provider::util::expand({ "a.b.c" = "value" }, ".")
}
# Result:
# { a = { b = { c = "value" } } }
