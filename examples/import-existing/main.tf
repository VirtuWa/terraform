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

resource "asashv_vm" "existing" {
  name        = "Test225"
  description = "None"

  vcpu      = 2
  vcores    = 1
  memory_gb = 4
  boot_type = "uefi"

  disks {
    pool           = "NFS-Storage"
    size_gb        = 20
    bus            = "virtio"
    format         = "qcow2"
    storage_target = "allocate"
  }

  cdrom {
    id = "L29wdC9zdG9yYWdlL25mcy9ORlMtU3RvcmFnZS91YnVudHUtMjYuMDQtbGl2ZS1zZXJ2ZXItYW1kNjQuaXNv"
  }

  nics = ["default"]

  boot_order = ["disk0", "cdrom0"]

  # Safer for imported production VMs.
  delete_disks = false
}
