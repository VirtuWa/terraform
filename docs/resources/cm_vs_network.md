# asashv_cm_vs_network

Creates and manages a VS/OVS network in LejamCM.

## Example

```hcl
resource "asashv_cm_vs_network" "test" {
  bridge_name  = "tf-vs-test"
  bond_mode    = "active-backup"
  cluster_name = "TF-HA-Cluster-Test"
  description  = "Terraform VS network test"

  host {
    hostname   = "VirtuWaHV-50"
    interfaces = ["enp5s0", "enp6s0"]
  }
}
```

## Arguments

- `bridge_name` - VS network bridge name.
- `bond_mode` - Bond mode. Default is `active-backup`.
- `cluster_name` - Cluster name where the VS network should be created.
- `description` - Description.
- `host.hostname` - Hostname where interfaces will be used.
- `host.interfaces` - Physical interfaces to use.

## Attributes

- `id` - VS network ID.
- `bond_name` - Bond name created by LejamCM.
- `cluster_id` - Resolved cluster ID.
- `status` - VS network status.
- `physical_uplinks` - Physical uplinks used by the VS network.
- `host.host_id` - Resolved host ID.

## Notes

The provider validates the VS network before creating it.

VS networks can affect host networking. Use only free interfaces.
