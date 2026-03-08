locals {
  features = {
    api     = true
    web     = false
    metrics = null
    cache   = "enabled"
  }
}

# Returns {api = true, web = false, cache = "enabled"} — nulls removed
output "compact_map" {
  value = provider::util::compact(local.features)
}
# Result:
# {
#   api   = true
#   cache = "enabled"
#   web   = false
# }
