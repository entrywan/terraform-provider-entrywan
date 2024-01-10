resource "entrywan_loadbalancer" "myloadbalancer" {
  name     = "myloadbalancer"
  location = "us1"
  protocol = "http"
  algo     = "round-robin"
  listeners {
    port = 80
    targets {
      ip   = "google.com"
      port = 80
    }
  }
}
