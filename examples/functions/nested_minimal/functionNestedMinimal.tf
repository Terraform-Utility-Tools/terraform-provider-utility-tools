locals {
  services = {
    dev  = { api = { port = 8080 }, web = null, cache = {} }
    prod = { api = { port = 8080 }, web = { port = 443 }, cache = {} }
  }
}

# Returns services with null and empty entries removed, preserving the nested structure
output "minimal_services" {
  value = provider::util::nestedMinimal(local.services)
}
# Result:
# {
#   dev  = { api = { port = 8080 } }
#   prod = { api = { port = 8080 }, web = { port = 443 } }
# }
