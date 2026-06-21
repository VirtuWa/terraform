# ASASHV Provider

The ASASHV provider manages virtual machines on ASASHV.

## Example

```hcl
provider "asashv" {
  endpoint = "https://192.168.1.220"
  username = var.asashv_username
  password = var.asashv_password
  insecure = true
}
```

## Environment Variables

- `ASASHV_ENDPOINT`
- `ASASHV_USERNAME`
- `ASASHV_PASSWORD`
- `ASASHV_TOKEN`
- `ASASHV_INSECURE`

## Resources

- `asashv_vm`

## Data Sources

- `asashv_iso`
- `asashv_storage_pool`
- `asashv_vlan`
