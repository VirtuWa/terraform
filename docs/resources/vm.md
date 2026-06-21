# asashv_vm

Manages an ASASHV virtual machine.

## Example

```hcl
resource "asashv_vm" "test" {
  name        = "tf-test-vm-001"
  description = "Created by Terraform"

  vcpu      = 1
  vcores    = 1
  memory_gb = 2
  boot_type = "uefi"

  disks {
    pool           = "NFS-Storage"
    size_gb        = 20
    bus            = "virtio"
    format         = "qcow2"
    storage_target = "allocate"
  }

  cdrom {
    id = data.asashv_iso.ubuntu.id
  }

  nics = ["default"]

  boot_order = ["disk0", "cdrom0"]

  delete_disks = true
}
```

## Import

```bash
terraform import asashv_vm.existing <vm-id>
```

For imported production VMs, use:

```hcl
delete_disks = false
```
