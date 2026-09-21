variable "name" {
  description = "Prefix for every resource name, e.g. fieldops-staging."
  type        = string
}

variable "env" {
  description = "Environment name; selects the SSM parameter path /fieldops/<env>/*."
  type        = string
}

variable "region" {
  description = "AWS region, used in the SSM parameter ARN and by the AWS CLI in user data. A variable rather than a data source so the module plans offline."
  type        = string
}

variable "vpc_id" {
  type = string
}

variable "public_subnet_ids" {
  description = "Subnets for the application load balancer."
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "Subnets for the backend instances."
  type        = list(string)
}

variable "ami_id" {
  description = "AMI for the backend instances. Expected to be Amazon Linux 2023 (or any AMI with dnf, the AWS CLI v2 and systemd): user data installs docker with dnf and reads SSM with the CLI."
  type        = string

  validation {
    condition     = can(regex("^ami-[0-9a-f]{8,17}$", var.ami_id))
    error_message = "ami_id must look like ami-0123456789abcdef0."
  }
}

variable "instance_type" {
  type    = string
  default = "t3.small"
}

variable "min_size" {
  type = number
}

variable "max_size" {
  type = number
}

variable "desired_capacity" {
  type = number
}

variable "backend_image" {
  description = "Full ECR image reference the instances run, e.g. 123456789012.dkr.ecr.eu-west-2.amazonaws.com/fieldops-backend:sha-abc123. The registry host is derived from it for docker login."
  type        = string
}

variable "ecr_repository_arn" {
  description = "ARN of the ECR repository that holds backend_image; the instance role may pull from this repository only."
  type        = string
}

variable "health_check_path" {
  description = "Path the target group probes on port 8080. The backend must answer 200 here without authentication."
  type        = string
  default     = "/api/health"
}

variable "seed_on_boot" {
  description = "Value of SEED_ON_BOOT passed to the backend container. Keep false anywhere the database is shared."
  type        = bool
  default     = false
}

variable "ssm_kms_key_arn" {
  description = "Customer-managed KMS key the /fieldops/<env>/* SecureStrings are encrypted with. Leave empty when they use the AWS-managed aws/ssm key, which needs no explicit kms:Decrypt grant."
  type        = string
  default     = ""
}
