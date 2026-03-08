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
  services_nested = {
    dev = {
      internal = {
        web = { port = 80 }
        api = { port = 8080 }
      }
      external = {
        web = { port = 443 }
        api = { port = 8443 }
      }
    }
    prod = {
      internal = {
        web = { port = 80 }
        api = { port = 8080 }
      }
      external = {
        web = { port = 443 }
        api = { port = 8443 }
      }
    }
  }
}

# Collapse 3 levels deep — leaves are the service objects
output "collapsed" {
  value = provider::util::collapse(local.services_nested, "/", 3)
}
