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

output "combine_with_separator" {
  value = provider::util::combine(local.input, "_")
}
