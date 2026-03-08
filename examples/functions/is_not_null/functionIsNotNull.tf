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
# Result:
# {
#   bool   = true
#   list   = true
#   map    = true
#   null   = false
#   number = true
#   string = true
# }
