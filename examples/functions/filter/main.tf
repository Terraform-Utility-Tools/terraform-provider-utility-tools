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
  services = {
    "dev/api"  = { env = "dev", port = 8080 }
    "dev/web"  = { env = "dev", port = 443 }
    "prod/api" = { env = "prod", port = 8080 }
  }
}

# Returns entries where env == "dev"
output "dev_services" {
  value = provider::util::filter(local.services, { env = "dev" })
}
# Result:
# {
#   "dev/api" = { env = "dev", port = 8080 }
#   "dev/web" = { env = "dev", port = 443 }
# }

# OR: dev OR prod
output "dev_or_prod" {
  value = provider::util::filter(local.services, { env = "dev" }, { env = "prod" })
}
# Result:
# {
#   "dev/api"  = { env = "dev", port = 8080 }
#   "dev/web"  = { env = "dev", port = 443 }
#   "prod/api" = { env = "prod", port = 8080 }
# }
