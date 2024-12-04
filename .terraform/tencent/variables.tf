variable "region" {
  description = "Tencent Cloud region"
  default     = "ap-beijing" 
}

variable "secret_id" {
  description = "Tencent Cloud secret ID"
}

variable "secret_key" {
  description = "Tencent Cloud secret key"
  sensitive   = true 
}
variable "tcr_username" {
    description = "Tencent Cloud Registry username"
}
variable "tcr_password" {
    description = "Tencent Cloud Registry password"
    sensitive   = true 
}
variable "container_image_name" {
    description = "Container image name"
    default     = "replay-api"
}
