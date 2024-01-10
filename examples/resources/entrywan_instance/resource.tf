resource "entrywan_instance" "myinstance" {
  hostname = "castula"
  location = "us1"
  disk     = 20
  cpus     = 1
  ram      = 2
  sshkey   = "mysshkey"
  os       = "debian"
}
