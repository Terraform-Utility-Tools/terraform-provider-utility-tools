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
  items = {
    string = "working"
    "null" = null
    map    = {}
    list   = []
    number = 0
    bool   = false
  }
}

output "is_not_null" {
  value = { for k, v in local.items : k => provider::util::isNotNull(v) }
}
