resource "hello_hot_air_balloon" "test" {
  name = "test-balloon"
  sku  = "standard"

  modification {
    id          = "satNav"
    description = "satellite navigation"
  }

  modification {
    id          = "satNav"
    description = "yet another sat nav modification, mistakenly added"
  }
}
