resource "entrywan_app" "nginx" {
  name     = "my-nginx-worker"
  location = "us1"
  image    = "nginx"
  port     = 80
  size     = 256
  source   = "oci"
}
