resource "hello_paper_plane" "test" {
  name = "test-paper-plane"
  # Remote API set sku to "standard" if not specified
}

resource "hello_hot_air_balloon" "test" {
  name                = "test-balloon"
  sku                 = hello_paper_plane.test.sku
  additional_gas_tank = 5
}
