# asashv_cm_host

Looks up a LejamCM host.

## Example

```hcl
data "asashv_cm_host" "hv50" {
  hostname = "VirtuWaHV-50"
}
```

## Arguments

- `hostname` - Hostname to search for.

## Attributes

- `id` - Host ID.
- `hostname` - Hostname.
- `ip_address` - Host IP address.
- `status` - Host status.
- `cluster_id` - Cluster ID if attached.
- `cluster_name` - Cluster name if attached.
