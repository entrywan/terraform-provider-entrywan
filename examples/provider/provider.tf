provider "entrywan" {
  token    = "iam_token_sensitive"
  endpoint = "https://api.entrywan.com/v1beta"
}

resource "entrywan_sshkey" "mysshkey" {
  name = "mysshkey"
  pub  = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBPdKY/JtRdXBoonecpczBwzGKSch8UIKGhLROjGLXBU root@betelgeuse"
}

resource "entrywan_instance" "castula" {
  hostname   = "castula"
  location   = "us1"
  disk       = 20
  cpus       = 1
  ram        = 2
  sshkey     = "mysshkey"
  os         = "debian"
  depends_on = entrywan_sshkey.mysshkey
}
