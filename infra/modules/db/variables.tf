variable "name" {
  description = "Prefix for every resource name, e.g. fieldops-staging."
  type        = string
}

variable "vpc_id" {
  type = string
}

variable "private_subnet_ids" {
  description = "Subnets for the DB subnet group; the instance is never publicly reachable."
  type        = list(string)
}

variable "compute_security_group_id" {
  description = "The only security group allowed to reach port 5432."
  type        = string
}

variable "instance_class" {
  type    = string
  default = "db.t4g.micro"
}

variable "allocated_storage" {
  description = "Initial storage in GiB; autoscales up to max_allocated_storage."
  type        = number
  default     = 20
}

variable "max_allocated_storage" {
  type    = number
  default = 100
}

variable "engine_version" {
  description = "PostgreSQL major version; minor upgrades are automatic."
  type        = string
  default     = "16"
}

variable "db_name" {
  type    = string
  default = "fieldops"
}

variable "username" {
  description = "Master username."
  type        = string
}

variable "password" {
  description = "Master password. Supplied at plan/apply time (TF_VAR_db_password or a gitignored tfvars); never written into this repository."
  type        = string
  sensitive   = true
}

variable "multi_az" {
  type    = bool
  default = false
}
