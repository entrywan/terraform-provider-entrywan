---
page_title: "Terraform Provider for Entrywan"
description: Manage [Entrywan](https://www.entrywan.com) resources with Terraform
---

# Entrywan Provider

The Entrywan provider allows managing [Entrywan](https://www.entrywan.com)
resources.  An active account and IAM token are required.

Available resources are described in this repository, and correspond
to their equivalents in the
[documentation](https://www.entrywan.com/docs).

## Example

The following example imports an ssh key and creates a compute
instance that uses that key.

```terraform
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
```

## Schema

### Required

- `endpoint` (String) Entrywan API endpoint
- `token` (String, Sensitive) Entrywan IAM token
