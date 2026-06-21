variable "asashv_endpoint" {
  type = string
}

variable "asashv_username" {
  type = string
}

variable "asashv_password" {
  type      = string
  sensitive = true
}

variable "asashv_insecure" {
  type    = bool
  default = true
}
