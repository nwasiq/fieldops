output "alb_dns_name" {
  description = "Backend entry point; the frontend's /api requests must reach this host."
  value       = module.compute.alb_dns_name
}

output "autoscaling_group_name" {
  description = "Target of the instance refresh that is a deploy."
  value       = module.compute.autoscaling_group_name
}

output "db_address" {
  value = module.db.address
}

output "db_port" {
  value = module.db.port
}

output "frontend_bucket_name" {
  value = module.edge.bucket_name
}

output "cloudfront_distribution_id" {
  value = module.edge.distribution_id
}

output "cloudfront_domain_name" {
  value = module.edge.distribution_domain_name
}

output "vpc_id" {
  value = module.network.vpc_id
}
