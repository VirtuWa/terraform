# asashv_cm_vlan

Creates and manages a VLAN/network in LejamCM.

## Example

```hcl
resource "asashv_cm_vlan" "vlan44" {
  vlan_id     = 44
  name        = "tf-vlan44"
  description = "Terraform test VLAN"
}
```

## Arguments

- `vlan_id` - VLAN ID.
- `name` - VLAN name.
- `description` - VLAN description.

## Attributes

- `id` - LejamCM VLAN resource ID.
- `bridge_name` - Bridge name used by LejamCM.
- `cluster_id` - Cluster ID if attached.
- `cluster_name` - Cluster name if attached.
- `primary_vs_network` - Primary VS network name.
- `vs_network_count` - Number of VS networks attached to the VLAN.

## Notes

The provider deletes VLANs using the real LejamCM VLAN ID, for example `44-br0`, and verifies deletion.
