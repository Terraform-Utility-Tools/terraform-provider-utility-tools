terraform {
  required_providers {
    util = {
      source = "Terraform-Utility-Tools/utility-tools"
    }
  }
  required_version = ">= 1.8.0"
}

provider "util" {}
