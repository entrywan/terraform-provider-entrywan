resource "entrywan_firewall" "myfirewall" {
  name = "http"

  rules {
    port     = "80"
    protocol = "tcp"
  }

  rules {
    port     = "443"
    protocol = "tcp"
  }
}
