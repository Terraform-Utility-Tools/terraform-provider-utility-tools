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
  entry = { env = "prod", region = "europe-west1", service = { id = "api", port = 8080 } }
}

output "without_region" {
  value = provider::util::omit(local.entry, ["region"])
}
# Result:
# {
#   env     = "prod"
#   service = { id = "api", port = 8080 }
# }
