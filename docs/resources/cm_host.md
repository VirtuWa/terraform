# asashv_cm_host

Adds and manages a host in LejamCM.

## Example

```hcl
variable "host_ssh_password" {
  type      = string
  sensitive = true
}

resource "asashv_cm_host" "node55" {
  ip_address   = "192.168.1.55"
  ssh_username = "admin"
  ssh_password = var.host_ssh_password
}
```

## Arguments

- `ip_address` - Host management IP address.
- `ssh_username` - SSH username used by LejamCM to add the host.
- `ssh_password` - SSH password used by LejamCM to add the host. This field is sensitive.

## Attributes

- `id` - LejamCM host ID.
- `hostname` - Hostname discovered by LejamCM.
- `cluster_id` - Cluster ID if the host belongs to a cluster.
- `cluster_name` - Cluster name if the host belongs to a cluster.
- `vendor` - Host vendor.
- `model` - Host model.
- `cpu_sockets` - CPU socket count.
- `cpu_cores_per_socket` - CPU cores per socket.
- `ram_gb` - Host RAM in GB.
- `cpu_usage_percent` - Current CPU usage percentage.
- `memory_usage_percent` - Current memory usage percentage.
- `status` - Host status.
- `maintenance_mode` - Whether the host is in maintenance mode.

## Notes

The host is created asynchronously. The provider waits until the host is no longer in `provisioning` state.
