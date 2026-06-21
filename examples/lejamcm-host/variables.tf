variable "lejamcm_endpoint" {
  type        = string
  description = "LejamCM endpoint, for example https://192.168.1.3."
}

variable "lejamcm_username" {
  type        = string
  description = "LejamCM username."
}

variable "lejamcm_password" {
  type        = string
  sensitive   = true
  description = "LejamCM password."
}

variable "lejamcm_insecure" {
  type        = bool
  default     = true
  description = "Allow self-signed/internal TLS certificates."
}

variable "host_ssh_password" {
  type        = string
  sensitive   = true
  description = "SSH password for the ASASHV host being added."
}
