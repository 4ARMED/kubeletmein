resource "digitalocean_kubernetes_cluster" "kubeletmein" {
  name   = var.cluster_name
  region = var.region_slug

  # Grab the latest version slug from `doctl kubernetes options versions`
  version = var.k8s_version

  node_pool {
    name       = "${var.cluster_name}-pool"
    size       = var.size
    node_count = var.node_count
  }
}