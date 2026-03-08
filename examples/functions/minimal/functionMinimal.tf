locals {
  config = {
    api     = { port = 8080 }
    web     = null
    metrics = {}
    cache   = []
  }
}

# Returns {api = {port = 8080}} — nulls and empty collections removed
output "minimal_config" {
  value = provider::util::minimal(local.config)
}
# Result:
# {
#   api = { port = 8080 }
# }
