locals {
  name = "fieldops-${var.env}"
}

module "network" {
  source = "../../modules/network"

  name                 = local.name
  vpc_cidr             = var.vpc_cidr
  availability_zones   = var.availability_zones
  public_subnet_cidrs  = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs
}

module "compute" {
  source = "../../modules/compute"

  name               = local.name
  env                = var.env
  region             = var.region
  vpc_id             = module.network.vpc_id
  public_subnet_ids  = module.network.public_subnet_ids
  private_subnet_ids = module.network.private_subnet_ids

  ami_id           = var.ami_id
  instance_type    = var.instance_type
  min_size         = var.asg_min_size
  max_size         = var.asg_max_size
  desired_capacity = var.asg_desired_capacity

  backend_image      = var.backend_image
  ecr_repository_arn = var.ecr_repository_arn
  health_check_path  = var.health_check_path
  seed_on_boot       = var.seed_on_boot
  ssm_kms_key_arn    = var.ssm_kms_key_arn
}

module "db" {
  source = "../../modules/db"

  name                      = local.name
  vpc_id                    = module.network.vpc_id
  private_subnet_ids        = module.network.private_subnet_ids
  compute_security_group_id = module.compute.instance_security_group_id

  instance_class    = var.db_instance_class
  allocated_storage = var.db_allocated_storage
  username          = var.db_username
  password          = var.db_password
  multi_az          = var.db_multi_az
}

module "edge" {
  source = "../../modules/edge"

  name        = local.name
  bucket_name = var.frontend_bucket_name
}
