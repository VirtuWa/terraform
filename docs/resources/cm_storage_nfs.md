# asashv_cm_storage_nfs

Creates and manages an NFS storage pool in LejamCM.

## Example

```hcl
resource "asashv_cm_storage_nfs" "nfs" {
  name            = "TF-NFS-Test"
  nfs_target_ip   = "192.168.1.37"
  nfs_remote_path = "/srv/nfsshare2"

  hostnames = [
    "VirtuWaHV-53"
  ]
}
```

## Arguments

- `name` - Storage pool name.
- `nfs_target_ip` - NFS server IP address.
- `nfs_remote_path` - NFS export path.
- `hostnames` - Hostnames where the NFS storage should be attached.

## Attributes

- `id` - Storage UUID.
- `host_ids` - Resolved LejamCM host IDs.
- `type` - Storage type.
- `protocol` - Storage protocol.
- `status` - Storage status.
- `size_gb` - Total size in GB.
- `used_gb` - Used size in GB.
- `free_gb` - Free size in GB.
- `host_count` - Number of attached hosts.

## Notes

Users provide hostnames. The provider resolves hostnames to LejamCM host IDs internally.

During destroy, the provider detaches hosts first, waits until `host_count = 0`, then deletes the storage pool.
