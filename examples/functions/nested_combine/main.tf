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
  input = [
    { key = "env", items = ["dev", "prod"] },
    { key = "region", items = ["europe-west1", "europe-north1"] },
    { key = "service", items = ["api", "web"] },
  ]
}

# Produces a nested map: env → region → service → {env, region, service}
output "nested_combine" {
  value = provider::util::nestedCombine(local.input)
}
# Result (excerpt):
# {
#   dev = {
#     europe-west1 = {
#       api = { env = "dev", region = "europe-west1", service = "api" }
#       web = { env = "dev", region = "europe-west1", service = "web" }
#     }
#     ...
#   }
#   ...
# }

output "nested_combine_with_separator" {
  value = provider::util::nestedCombine(local.input, "_")
}
# Result (excerpt): same structure, keys joined with "_" at each level
