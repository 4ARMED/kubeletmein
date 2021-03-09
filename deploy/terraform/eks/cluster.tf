module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  cluster_name    = var.cluster_name
  cluster_version = "1.18"
  subnets         = module.vpc.private_subnets

  tags = {
    GithubRepo  = "kubeletmein"
    GithubOrg   = "4armed"
  }

  vpc_id = module.vpc.vpc_id

  workers_group_defaults = {
    root_volume_type = "gp2"
  }

  worker_groups = [
    {
      name                          = "${var.cluster_name}-worker-group"
      instance_type                 = var.instance_type
      asg_desired_capacity          = var.asg_desired_capacity
    }
  ]
}