data "google_client_config" "provider" {}

data "google_container_cluster" "kubeletmein" {
  name     = google_container_cluster.kubeletmein.name
}

provider "kubernetes" {
  host  = "https://${data.google_container_cluster.kubeletmein.endpoint}"
  token = data.google_client_config.provider.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.kubeletmein.master_auth[0].cluster_ca_certificate,
  )
}

resource "kubernetes_pod" "kubeletmein" {
  depends_on = [ google_container_node_pool.kubeletmein_vulnerable_nodes ]

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

    node_selector = {
      "vulnerable" = "true"
    }
  }

  timeouts {
    create = "30m" // Allow enough time for the cluster to be happy
    delete = "10m"
  }
}