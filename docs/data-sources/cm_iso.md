# asashv_cm_iso

Looks up an ISO file in LejamCM.

## Example

```hcl
data "asashv_cm_iso" "ubuntu" {
  name = "ubuntu-26.04-live-server-amd64.iso"
}
```

## Arguments

- `name` - ISO file name.

## Attributes

- `id` - ISO ID.
- `name` - ISO name.
- `size_gb` - ISO size in GB.
