# Terraform Provider ASASHV

Terraform provider for ASASHV hypervisor management.

This provider allows Terraform to manage ASASHV virtual machines and read ASASHV storage, ISO, and VLAN/network information.

## Current Features

### Provider Configuration

- `endpoint`
- `username`
- `password`
- `token`
- `insecure`

### Resources

- `asashv_vm`

### Data Sources

- `asashv_iso`
- `asashv_storage_pool`
- `asashv_vlan`

## Example

```hcl
terraform {
  required_providers {
    asashv = {
      source  = "asashv/asashv"
      version = "0.1.0"
    }
  }
}

provider "asashv" {
  endpoint = "https://192.168.1.220"
  username = var.asashv_username
  password = var.asashv_password
  insecure = true
}

data "asashv_storage_pool" "nfs" {
  name = "NFS-Storage"
}

data "asashv_iso" "ubuntu" {
  name = "ubuntu-26.04-live-server-amd64.iso"
}

data "asashv_vlan" "default" {
  name = "default"
}

resource "asashv_vm" "test" {
  name        = "tf-test-vm-001"
  description = "Created by Terraform"

  vcpu      = 1
  vcores    = 1
  memory_gb = 2
  boot_type = "uefi"

  disks {
    pool           = data.asashv_storage_pool.nfs.name
    size_gb        = 20
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
