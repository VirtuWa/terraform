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

resource "asashv_cm_storage_nfs" "test" {
  name            = "TF-NFS-Test"
  nfs_target_ip   = "192.168.1.37"
  nfs_remote_path = "/srv/nfsshare2"

  hostnames = [
    "VirtuWaHV-53"
  ]
}

output "storage_id" {
  value = asashv_cm_storage_nfs.test.id
}

output "storage_status" {
  value = asashv_cm_storage_nfs.test.status
}

output "storage_host_count" {
  value = asashv_cm_storage_nfs.test.host_count
}

output "storage_host_ids" {
  value = asashv_cm_storage_nfs.test.host_ids
}

output "storage_hostnames" {
  value = asashv_cm_storage_nfs.test.hostnames
}
