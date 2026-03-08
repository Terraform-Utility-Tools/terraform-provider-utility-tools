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
    { key = "env", items = ["dev", "stage", "prod"] },
    { key = "region", items = ["europe-west1", "europe-north1"] },
    { key = "service", items = {
      default = { neg = "default", value = 100 }
      api     = { neg = "api", override = 200 }
    } },
  ]
}

output "combine" {
  value = provider::util::combine(local.input)
}
# Result (excerpt):
# {
#   "dev/europe-west1/api"     = { env = "dev", region = "europe-west1", service = { neg = "api", override = 200 }, id = "api" }
#   "dev/europe-west1/default" = { env = "dev", region = "europe-west1", service = { neg = "default", value = 100 }, id = "default" }
#   ...
# }

output "combine_with_separator" {
  value = provider::util::combine(local.input, "_")
}
# Result (excerpt):
# {
#   "dev_europe-west1_api" = { env = "dev", region = "europe-west1", service = { neg = "api", override = 200 }, id = "api" }
#   ...
# }
