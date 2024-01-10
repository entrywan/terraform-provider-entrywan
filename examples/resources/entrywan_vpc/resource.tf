resource "entrywan_vpc" "vpc" {
  name   = "myvpc"
  prefix = "192.168.5.0/24"
  members {
    ip4public = "38.22.213.9"
  }
  members {
    ip4public = "38.95.213.54"
    ip4private = "192.168.5.122"
  }
}
