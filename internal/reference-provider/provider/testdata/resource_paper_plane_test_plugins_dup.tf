resource "hello_paper_plane" "test" {
  name = "test-paper-plane"

  plugins {
    id          = "satnav"
    description = "Navigate the paper plane with GPS"
  }

  plugins {
    id          = "satnav"
    description = "Mistakenly duplicated satnav"
  }
}
