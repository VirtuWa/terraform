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

resource "asashv_cm_vs_network" "test" {
  bridge_name  = "tf-vs-test"
  bond_mode    = "active-backup"
  cluster_name = "TF-HA-Cluster-Test"
  description  = "Terraform VS network test"

  host {
    hostname   = "VirtuWaHV-50"
    interfaces = ["enp5s0", "enp6s0"]
  }
}

output "vs_network_id" {
  value = asashv_cm_vs_network.test.id
}

output "vs_network_status" {
  value = asashv_cm_vs_network.test.status
}

output "vs_network_bond_name" {
  value = asashv_cm_vs_network.test.bond_name
}

output "vs_network_cluster_id" {
  value = asashv_cm_vs_network.test.cluster_id
}

output "vs_network_physical_uplinks" {
  value = asashv_cm_vs_network.test.physical_uplinks
}
