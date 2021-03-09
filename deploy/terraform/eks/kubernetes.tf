data "aws_eks_cluster" "kubeletmein" {
  name = module.eks.cluster_id
}

data "aws_eks_cluster_auth" "kubeletmein" {
  name = module.eks.cluster_id
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.kubeletmein.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.kubeletmein.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.kubeletmein.token
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

  timeouts {
    create = "30m" // Allow enough time for the cluster to be happy
    delete = "10m"
  }
}