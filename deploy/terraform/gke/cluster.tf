resource "google_container_cluster" "kubeletmein" {
  name               = var.cluster_name
  description = "Example GKE configuration for testing 4ARMED's kubeletmein tool"
  location = var.location

    # Remove default node pool and create one with two 
  # preemptible nodes for $$$ savings
  remove_default_node_pool = true
  initial_node_count = 1

  release_channel {
    channel = "REGULAR"
  }
}

resource "google_container_node_pool" "kubeletmein_vulnerable_nodes" {
  name = "${var.cluster_name}-vulnerable-node-pool"
  cluster = google_container_cluster.kubeletmein.name
  node_count = 1
  location = var.location
  
  node_config {
    preemptible = true
    machine_type = var.machine_type

    labels = {
      vulnerable = "true"
    }
  }

  timeouts {
    create = "30m"
    update = "40m"
  }
}