locals {
  user_data = templatefile("cloud-init.yaml", {
    user  = var.USER
  })
}

resource "hcloud_server" "server" {
  name         = "server"
  image        = "debian-12"
  server_type  = "cax21"
  location     = "fsn1"
  user_data    = local.user_data

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
}

variable "USER" {}
