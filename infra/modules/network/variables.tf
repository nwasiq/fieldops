variable "name" {
  description = "Prefix for every resource name, e.g. fieldops-staging."
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
}

variable "availability_zones" {
  description = "Exactly two availability zones; one public and one private subnet is created in each. A list, never a data source, so the module plans without AWS access."
  type        = list(string)

  validation {
    condition     = length(var.availability_zones) == 2
    error_message = "availability_zones must list exactly two zones."
  }
}

variable "public_subnet_cidrs" {
  description = "One CIDR per availability zone for the public subnets (ALB, NAT gateway)."
  type        = list(string)

  validation {
    condition     = length(var.public_subnet_cidrs) == 2
    error_message = "public_subnet_cidrs must list exactly two CIDRs, one per availability zone."
  }
}

variable "private_subnet_cidrs" {
  description = "One CIDR per availability zone for the private subnets (backend instances, RDS)."
  type        = list(string)

  validation {
    condition     = length(var.private_subnet_cidrs) == 2
    error_message = "private_subnet_cidrs must list exactly two CIDRs, one per availability zone."
  }
}
