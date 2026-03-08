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
  yaml_files = fileset(path.module, "config/*.{yaml,yml}")
  configs = {
    for f in local.yaml_files :
    provider::util::trimext(basename(f)) => yamldecode(file("${path.module}/${f}"))
  }
}

output "configs" {
  value = local.configs
}
# Result (given config/app.yaml, config/database.yml):
# {
#   "app"      = { port = 8080, replicas = 2 }
#   "database" = { host = "db.internal", port = 5432 }
# }
