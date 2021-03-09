variable "region" {
  type = string
  default = "us-east-1"
}

variable "cidr_block" {
  type = string
  default = "10.99.0.0/16"
}

variable "public_subnets" {
  type = list
  default = ["10.99.0.0/24", "10.99.1.0/24"]
}

variable "private_subnets" {
  type = list
  default = ["10.99.100.0/24", "10.99.101.0/24"]
}

variable "cluster_name" {
  type = string
  default = "kubeletmein"
}

variable "instance_type" {
  type = string
  default = "t2.small"
}

variable "asg_desired_capacity" {
  type = number
  default = 1
}

variable "kubeletmein_image" {
  type = string
  default = "4armed/kubeletmein:latest"
}