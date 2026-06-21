# asashv_cm_vm

Creates and manages a VM through LejamCM.

## Example

```hcl
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
```

## Arguments

- `name` - VM name.
- `description` - VM description.
- `host_id` - LejamCM host ID where the VM should be created.
- `vcpu` - Number of virtual CPUs.
- `vcores` - Number of cores.
- `memory_gb` - VM memory in GB.
- `boot_type` - Boot type, for example `uefi`.
- `disks` - VM disk configuration.
- `cdrom` - ISO/CD-ROM configuration.
- `nics` - Network names/VLAN names.
- `boot_order` - Boot order list.
- `delete_disks` - Whether disks should be deleted when the VM is destroyed.

## Attributes

- `id` - VM ID.
- `ip` - VM IP address if discovered.
- `power_state` - Current VM power state.
- `os` - Detected operating system.
- `host_name` - Host where the VM exists.
- `host_ip` - Host IP.
- `cluster_name` - Cluster name.
- `unmount_cdrom` - CD-ROM unmount status.
