data "digitalocean_kubernetes_cluster" "kubeletmein" {
  name = var.cluster_name

  depends_on = [ digitalocean_kubernetes_cluster.kubeletmein ]
}

provider "kubernetes" {
  host             = data.digitalocean_kubernetes_cluster.kubeletmein.endpoint
  token            = data.digitalocean_kubernetes_cluster.kubeletmein.kube_config[0].token
  cluster_ca_certificate = base64decode(
    data.digitalocean_kubernetes_cluster.kubeletmein.kube_config[0].cluster_ca_certificate
  )
}

resource "kubernetes_pod" "kubeletmein" {

  metadata {
    name = "kubeletmein-vulnerable"
  }

  spec {
    container {
      image = var.kubeletmein_image
      name = "kubeletmein"
      command = [ "/bin/sleep", "99d"]

      # Attach to this pod with:
      # kubectl exec -ti kubeletmein-vulnerable bash
    }
  }
}