output "subnet_ids" {
  description = "IDs of the private subnets for ECS tasks"
  value       = [aws_subnet.private_1.id, aws_subnet.private_2.id]
}

output "security_group_id" {
  description = "Security group ID for ECS"
  value       = aws_security_group.this.id
}

output "target_group_arn" {
  description = "ARN of the ALB target group"
  value       = aws_lb_target_group.main.arn
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.main.dns_name
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "IDs of the public subnets for ALB"
  value       = [aws_subnet.public_1.id, aws_subnet.public_2.id]
}