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

resource "asashv_cm_vlan" "test" {
  vlan_id     = 44
  name        = "tf-vlan44"
  description = "Terraform test VLAN"
}

output "vlan_id" {
  value = asashv_cm_vlan.test.id
}

output "vlan_bridge" {
  value = asashv_cm_vlan.test.bridge_name
}

output "vlan_primary_vs_network" {
  value = asashv_cm_vlan.test.primary_vs_network
}
