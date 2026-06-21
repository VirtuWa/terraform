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

data "asashv_cm_host" "hv50" {
  hostname = "VirtuWaHV-50"
}

data "asashv_cm_storage" "nfs" {
  name = "NFS"
}

data "asashv_cm_iso" "ubuntu" {
  name = "ubuntu-26.04-live-server-amd64.iso"
}

resource "asashv_cm_vm" "test" {
  name        = "tf-cm-test-vm-001"
  description = "Created by Terraform through LejamCM"

  host_id   = data.asashv_cm_host.hv50.id
  vcpu      = 1
  vcores    = 1
  memory_gb = 2
  boot_type = "uefi"

  disks {
    pool           = data.asashv_cm_storage.nfs.name
    size_gb        = 20
    bus            = "virtio"
    format         = "qcow2"
    storage_target = "allocate"
  }

  cdrom {
    id = data.asashv_cm_iso.ubuntu.id
  }

  nics = ["default"]

  boot_order = ["disk0", "cdrom0"]

  delete_disks = true
}

output "cm_vm_id" {
  value = asashv_cm_vm.test.id
}

output "cm_vm_power_state" {
  value = asashv_cm_vm.test.power_state
}

output "cm_vm_host" {
  value = asashv_cm_vm.test.host_name
}
