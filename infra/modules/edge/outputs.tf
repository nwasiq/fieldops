output "bucket_name" {
  value = aws_s3_bucket.frontend.bucket
}

output "distribution_id" {
  description = "For the cache invalidation that follows a frontend upload."
  value       = aws_cloudfront_distribution.frontend.id
}

output "distribution_domain_name" {
  value = aws_cloudfront_distribution.frontend.domain_name
}
