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
  defaults  = { replicas = 1, tls = false, labels = { app = "myapp" } }
  overrides = { tls = true, labels = { env = "prod" } }
}

output "merged" {
  value = provider::util::nestedMerge(local.defaults, local.overrides)
}
# Result:
# {
#   replicas = 1
#   tls      = true
#   labels   = { app = "myapp", env = "prod" }
# }
