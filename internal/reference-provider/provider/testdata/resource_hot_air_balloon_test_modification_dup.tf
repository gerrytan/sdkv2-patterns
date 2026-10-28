resource "hello_hot_air_balloon" "test" {
  name = "test-balloon"
  sku  = "standard"

  modification { # this element will be overwritten by the one underneath because it has the same id!
    id          = "satNav"
    description = "satellite navigation"
  }

  modification {
    id          = "satNav"
    description = "yet another sat nav modification, mistakenly added"
  }
}
