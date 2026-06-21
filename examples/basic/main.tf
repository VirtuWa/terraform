terraform {
  required_providers {
    asashv = {
      source  = "asashv/asashv"
      version = "0.1.0"
    }
  }
}

provider "asashv" {
  endpoint = var.asashv_endpoint
  username = var.asashv_username
  password = var.asashv_password
  insecure = var.asashv_insecure
}

data "asashv_storage_pool" "nfs" {
  name = var.storage_pool_name
}

data "asashv_iso" "ubuntu" {
  name = var.iso_name
}

data "asashv_vlan" "default" {
  name = var.network_name
}

resource "asashv_vm" "test" {
  name        = var.vm_name
  description = "Created by Terraform"

  vcpu      = var.vcpu
  vcores    = var.vcores
  memory_gb = var.memory_gb
  boot_type = var.boot_type

  disks {
    pool           = data.asashv_storage_pool.nfs.name
    size_gb        = var.disk_size_gb
    bus            = "virtio"
    format         = "qcow2"
    storage_target = "allocate"
  }

  cdrom {
    id = data.asashv_iso.ubuntu.id
  }

  nics = [data.asashv_vlan.default.name]

  boot_order = ["disk0", "cdrom0"]

  delete_disks = true
}
