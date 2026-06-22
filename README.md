# Terraform Provider for AsasHV and Lejam CM

Terraform provider for managing infrastructure resources on **AsasHV Hypervisor** and **Lejam CM**.

## Supported Platforms

- AsasHV Hypervisor
- Lejam CM

## Features

- Virtual machine management
- Cluster management
- Host management
- Storage management
- VLAN management
- Virtual switch network management
- ISO discovery
- Storage pool discovery

## Provider Configuration

The provider supports the following configuration options:

- `endpoint`
- `username`
- `password`
- `token`
- `insecure`

## Resources

### AsasHV

- `asashv_vm`

### Lejam CM

- `cm_cluster`
- `cm_host`
- `cm_storage_nfs`
- `cm_vlan`
- `cm_vm`
- `cm_vs_network`

## Data Sources

### AsasHV

- `asashv_iso`
- `asashv_storage_pool`
- `asashv_vlan`

### Lejam CM

- `cm_host`
- `cm_iso`
- `cm_storage`

## Example

```hcl
terraform {
  required_providers {
    asashv = {
      source  = "VirtuWa/asashv"
      version = "0.1.0"
    }
  }
}

provider "asashv" {
  endpoint = "https://192.168.1.220"
  username = var.asashv_username
  password = var.asashv_password
  insecure = true
}
