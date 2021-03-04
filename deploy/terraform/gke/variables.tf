variable "project_id" {
  type = string
}

variable "location" {
  type = string
  default = "us-central1-a"
}

variable "cluster_name" {
  type = string
  default = "kubeletmein"
}

variable "machine_type" {
  type = string
  default = "g1-small"
}

variable "kubeletmein_image" {
  type = string
  default = "4armed/kubeletmein:latest"
}