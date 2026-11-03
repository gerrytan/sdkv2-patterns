resource "hello_paper_plane" "test" {
  name = "test-paper-plane"

  plugins {
    id          = "b"
    description = "Plugin b"
  }

  plugins {
    id          = "a"
    description = "Plugin a"
  }
}
