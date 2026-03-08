terraform {
  required_providers {
    util = {
      source = "Terraform-Utility-Tools/utility-tools"
    }
  }
  required_version = ">= 1.8.0"
}

# Default configuration — uses "/" as the separator for collapse, expand, and combine.
provider "util" {
}

# Custom separator — use "." as the default separator for all path-based functions.
provider "util" {
  separator = "."
}
