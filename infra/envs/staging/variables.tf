variable "env" {
  description = "Environment name; prefixes resource names and selects /fieldops/<env>/* in SSM."
  type        = string
  default     = "staging"

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]*$", var.env))
    error_message = "env must be lower-case letters, digits and hyphens."
  }
}

variable "region" {
  type = string
}

# --- network ---------------------------------------------------------------

variable "vpc_cidr" {
  type = string
}

variable "availability_zones" {
  description = "Exactly two zones in var.region. A list, not a data source: the plan must run offline."
  type        = list(string)
}

variable "public_subnet_cidrs" {
  type = list(string)
}

variable "private_subnet_cidrs" {
  type = list(string)
}

# --- compute ---------------------------------------------------------------

variable "ami_id" {
  description = "Amazon Linux 2023 AMI id for var.region."
  type        = string
}

variable "instance_type" {
  type    = string
  default = "t3.small"
}

variable "asg_min_size" {
  type = number
}

variable "asg_max_size" {
  type = number
}

variable "asg_desired_capacity" {
  type = number
}

variable "backend_image" {
  description = "ECR image reference the backend instances run."
  type        = string
}

variable "ecr_repository_arn" {
  description = "ARN of the ECR repository holding backend_image."
  type        = string
}

variable "health_check_path" {
  type    = string
  default = "/api/health"
}

variable "seed_on_boot" {
  type    = bool
  default = false
}

variable "ssm_kms_key_arn" {
  description = "Customer-managed KMS key for the SSM SecureStrings; empty for the AWS-managed aws/ssm key."
  type        = string
  default     = ""
}

# --- db --------------------------------------------------------------------

variable "db_instance_class" {
  type    = string
  default = "db.t4g.micro"
}

variable "db_allocated_storage" {
  type    = number
  default = 20
}

variable "db_username" {
  type = string
}

variable "db_password" {
  description = "RDS master password. Set via TF_VAR_db_password or a gitignored tfvars; the .example carries a placeholder only."
  type        = string
  sensitive   = true
}

variable "db_multi_az" {
  type    = bool
  default = false
}

# --- edge ------------------------------------------------------------------

variable "frontend_bucket_name" {
  description = "Globally unique S3 bucket name for the built frontend."
  type        = string
}
