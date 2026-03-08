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

output "is_null" {
  value = { for k, v in local.items : k => provider::util::isNull(v) }
}
# Result:
# {
#   bool   = false
#   list   = false
#   map    = false
#   null   = true
#   number = false
#   string = false
# }
