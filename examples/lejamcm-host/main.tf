terraform {
  required_providers {
    asashv = {
      source  = "asashv/asashv"
      version = "0.1.0"
    }
  }
}

provider "asashv" {
  endpoint = var.lejamcm_endpoint
  username = var.lejamcm_username
  password = var.lejamcm_password
  insecure = var.lejamcm_insecure
}

resource "asashv_cm_host" "node55" {
  ip_address   = "192.168.1.55"
  ssh_username = "admin"
  ssh_password = var.host_ssh_password
}

output "host_id" {
  value = asashv_cm_host.node55.id
}

output "hostname" {
  value = asashv_cm_host.node55.hostname
}

output "host_status" {
  value = asashv_cm_host.node55.status
}

output "host_ram_gb" {
  value = asashv_cm_host.node55.ram_gb
}
