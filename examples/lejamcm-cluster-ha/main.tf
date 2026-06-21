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

resource "asashv_cm_cluster" "test" {
  name                = "TF-HA-Cluster-Test"
  description         = "Terraform HA test cluster"
  virtual_ip          = "192.168.1.240"
  shared_storage_name = "NFS"

  hostnames = [
    "VirtuWaHV-50",
    "VirtuWaHV-53"
  ]

  ha_enabled  = true
  drs_enabled = false
}

output "cluster_id" {
  value = asashv_cm_cluster.test.id
}

output "cluster_status" {
  value = asashv_cm_cluster.test.status
}

output "cluster_ha_enabled" {
  value = asashv_cm_cluster.test.ha_enabled
}

output "cluster_host_count" {
  value = asashv_cm_cluster.test.host_count
}

output "cluster_host_ids" {
  value = asashv_cm_cluster.test.host_ids
}

output "cluster_hostnames" {
  value = asashv_cm_cluster.test.hostnames
}
