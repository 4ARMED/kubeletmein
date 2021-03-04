variable "do_token" {}

variable "cluster_name" {
  type = string
  default = "kubeletmein"
}

variable "region_slug" {
  type = string
  default = "lon1"
}

variable "k8s_version" {
  type = string
  default = "1.20.2-do.0"
}

variable "size" {
  type = string
  # doctl compute size list
  default = "s-1vcpu-2gb"
}

variable "node_count" {
  type = number
  # doctl compute size list
  default = 2
}

variable "kubeletmein_image" {
  type = string
  default = "4armed/kubeletmein:latest"
}