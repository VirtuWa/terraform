# asashv_cm_cluster

Creates and manages a LejamCM HA cluster.

## Example

```hcl
resource "asashv_cm_cluster" "ha" {
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
```

## Arguments

- `name` - Cluster name.
- `description` - Cluster description.
- `virtual_ip` - Cluster virtual IP.
- `shared_storage_name` - Shared storage pool name.
- `hostnames` - Hostnames to add to the cluster. LejamCM requires at least 2 hosts.
- `ha_enabled` - Whether HA should be enabled.
- `drs_enabled` - Whether DRS should be enabled.

## Attributes

- `id` - Cluster ID.
- `host_ids` - Resolved host IDs.
- `shared_storage_id` - Resolved shared storage ID.
- `status` - Cluster status.
- `host_count` - Number of hosts in the cluster.
- `vm_count` - Number of VMs in the cluster.
- `cpu_usage_percent` - CPU usage percentage.
- `memory_usage_percent` - Memory usage percentage.

## Notes

Users provide hostnames and shared storage name. The provider resolves them internally.

After creation, LejamCM may return `provisioning` first, then later change to `active`.
