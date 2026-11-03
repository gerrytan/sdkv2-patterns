resource "hello_paper_plane" "test" {
  name = "test-paper-plane"

  plugins_v2 {
    id          = "b"
    description = "Plugin b"
  }

  plugins_v2 {
    id          = "a"
    description = "Plugin a"
  }
}

resource "hello_paper_plane" "test2" {
  name = "test-paper-plane"

  plugins_v2 {
    id          = hello_paper_plane.test.plugins_v2[0].id
    description = hello_paper_plane.test.plugins_v2[0].description
  }
}