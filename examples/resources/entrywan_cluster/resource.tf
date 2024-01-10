resource "entrywan_cluster" "mycluster" {
  hostname = "mycluster"
  location = "us1"
  size     = 3
  cni      = "flannel"
}

