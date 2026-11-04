resource "hello_paper_plane" "test" {
  name = "test-paper-plane-stickers"

  stickers {
    id          = "sunrise"
    description = "A golden sunrise"
  }

  stickers {
    id          = "stars"
    description = "A sky full of stars"
  }
}
