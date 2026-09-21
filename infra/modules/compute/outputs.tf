output "alb_dns_name" {
  value = aws_lb.this.dns_name
}

output "alb_arn" {
  value = aws_lb.this.arn
}

output "instance_security_group_id" {
  description = "Security group of the backend instances; the db module admits it and nothing else."
  value       = aws_security_group.instance.id
}

output "autoscaling_group_name" {
  value = aws_autoscaling_group.backend.name
}

output "instance_role_name" {
  value = aws_iam_role.instance.name
}
