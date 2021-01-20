terraform {
  backend "s3" {
    bucket         = "terraform-state-storage-586877430255"
    dynamodb_table = "terraform-state-lock-586877430255"
    region         = "us-west-2"

    // THIS MUST BE UNIQUE
    key = "steve.tfstate"
  }
}

provider "aws" {
  region = "us-west-2"
}

data "aws_ssm_parameter" "eks_cluster_endpoint" {
  name = "/eks/av-cluster-endpoint"
}

provider "kubernetes" {
  host = data.aws_ssm_parameter.eks_cluster_endpoint.value
}

module "steve" {
  source = "github.com/byuoitav/terraform//modules/kubernetes-deployment"

  // required
  name           = "steve"
  image          = "docker.pkg.github.com/byuoitav/steve/steve-dev"
  image_version  = ""
  container_port = 8080
  repo_url       = "https://github.com/byuoitav/steve"

  // optional
  image_pull_secret = "github-docker-registry"
  container_env     = {}
  container_args = [
    //"--db-address", data.aws_ssm_parameter.prd_db_addr.value,
    //"--db-username", data.aws_ssm_parameter.prd_db_username.value,
    //"--db-password", data.aws_ssm_parameter.prd_db_password.value,
    //"--hub-address", data.aws_ssm_parameter.hub_address.value,
  ]
  health_check = false
}
