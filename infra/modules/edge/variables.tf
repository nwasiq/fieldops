variable "name" {
  description = "Prefix for every resource name, e.g. fieldops-staging."
  type        = string
}

variable "bucket_name" {
  description = "Globally unique S3 bucket name for the built frontend."
  type        = string
}

variable "price_class" {
  type    = string
  default = "PriceClass_100"
}
