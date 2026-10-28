resource "hello_hot_air_balloon" "test" {
  name = "test-balloon"
  sku  = "standard"

  modification_v2 {
    id          = "satNav"
    description = "satellite navigation"
  }

  modification_v2 {
    id          = "satNav"
    description = "duplicate satellite navigation"
  }
}
