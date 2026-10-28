resource "hello_hot_air_balloon" "test" {
  name            = "test-balloon"
  sku             = "standard"
  exterior_colour = ["red", "blue"]
}

resource "hello_paper_plane" "test" {
  name            = "test-paper-plane"
  exterior_colour = hello_hot_air_balloon.test.exterior_colour[1]
}
