# asashv_cm_storage

Looks up a LejamCM storage pool.

## Example

```hcl
data "asashv_cm_storage" "nfs" {
  name = "NFS"
}
```

## Arguments

- `name` - Storage pool name.

## Attributes

- `id` - Storage UUID.
- `name` - Storage name.
- `status` - Storage status.
- `type` - Storage type.
- `protocol` - Storage protocol.
- `size_gb` - Total size in GB.
- `used_gb` - Used size in GB.
- `free_gb` - Free size in GB.
