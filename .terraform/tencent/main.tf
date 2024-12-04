terraform {
  required_providers {
    tencentcloud = {
      source  = "tencentcloudstack/tencentcloud"
      version =
    }
  }
}

provider "tencentcloud" {
  region = var.region 
  secret_id  = var.secret_id
  secret_key = var.secret_key
}

resource "tencentcloud_api_gateway_api" "replay_api" {
  # ... API gateway configuration (paths, methods, authentication, etc.) ...
}

resource "tencentcloud_vpc" "replay_api_vpc" {
  # ... VPC configuration ...
}

resource "tencentcloud_subnet" "replay_api_subnet" {
  # ... Subnet configuration ...
}

resource "tencentcloud_mongodb_instance" "replay_api_mongodb" {
  # ... MongoDB instance configuration ...
}
resource "tencentcloud_security_group" "replay_api" {
  # ... Security group rules ...
}

resource "tencentcloud_container_cluster" "replay_api_cluster" {
  # ... TKE Kubernetes cluster configuration ...
}

data "tencentcloud_container_registry_repository" "replay_api_repository" {
  name = var.container_image_name
}

resource "tencentcloud_container_instance" "replay_api" {
  # ... Container instance configuration (image, environment, etc.) ...
  image_registry_credentials {
    name     = data.tencentcloud_container_registry_repository.replay_api_repository.name
    server   = "ccr.${var.region}.tencentcloudcr.com" 
    username = var.tcr_username
    password = var.tcr_password
  }
}

