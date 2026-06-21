variable "asashv_endpoint" {
  type        = string
  description = "ASASHV endpoint, for example https://192.168.1.220."
}

variable "asashv_username" {
  type        = string
  description = "ASASHV username."
}

variable "asashv_password" {
  type        = string
  sensitive   = true
  description = "ASASHV password."
}

variable "asashv_insecure" {
  type        = bool
  default     = true
  description = "Allow self-signed/internal TLS certificates."
}

variable "vm_name" {
  type    = string
  default = "tf-test-vm-001"
}

variable "vcpu" {
  type    = number
  default = 1
}

variable "vcores" {
  type    = number
  default = 1
}

variable "memory_gb" {
  type    = number
  default = 2
}

variable "boot_type" {
  type    = string
  default = "uefi"
}

variable "disk_size_gb" {
  type    = number
  default = 20
}

variable "storage_pool_name" {
  type    = string
  default = "NFS-Storage"
}

variable "iso_name" {
  type    = string
  default = "ubuntu-26.04-live-server-amd64.iso"
}

variable "network_name" {
  type    = string
  default = "default"
}
