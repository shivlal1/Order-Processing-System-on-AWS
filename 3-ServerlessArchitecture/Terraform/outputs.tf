output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs_receiver.cluster_name
}

output "receiver_service_name" {
  description = "Name of the order receiver ECS service"
  value       = module.ecs_receiver.service_name
}

# NEW: Lambda outputs
output "lambda_function_name" {
  description = "Name of the Lambda processor function"
  value       = module.lambda_processor.function_name
}

output "lambda_function_url" {
  description = "AWS Console URL for Lambda function"
  value       = module.lambda_processor.function_url
}

# ALB endpoint to access the service
output "load_balancer_url" {
  description = "URL of the Application Load Balancer"
  value       = "http://${module.network.alb_dns_name}"
}

# Sync endpoint (Phase 1 - slow)
output "test_sync_endpoint" {
  description = "Synchronous order processing (3s delay)"
  value       = "curl -X POST http://${module.network.alb_dns_name}/orders/sync -H 'Content-Type: application/json' -d '{\"customer_id\": 123, \"items\": [{\"product_id\": \"prod-1\", \"quantity\": 2, \"price\": 29.99}]}'"
}

# Async endpoint (Lambda - fast)
output "test_async_endpoint" {
  description = "Asynchronous order processing via Lambda (<100ms)"
  value       = "curl -X POST http://${module.network.alb_dns_name}/orders/async -H 'Content-Type: application/json' -d '{\"customer_id\": 456, \"items\": [{\"product_id\": \"prod-2\", \"quantity\": 1, \"price\": 49.99}]}'"
}

output "health_check_endpoint" {
  description = "Health check endpoint"
  value       = "http://${module.network.alb_dns_name}/health"
}

output "stats_endpoint" {
  description = "Stats endpoint"
  value       = "http://${module.network.alb_dns_name}/stats"
}

# AWS Resources for monitoring
output "sns_topic_arn" {
  description = "ARN of the SNS topic"
  value       = module.sns.topic_arn
}

# Lambda monitoring instructions
output "lambda_monitoring" {
  description = "How to monitor Lambda performance"
  value       = <<EOT
Monitor Lambda performance:
1. CloudWatch Logs: ${module.lambda_processor.log_group_name}
2. Look for "REPORT" lines to see:
   - Duration: Actual execution time
   - Billed Duration: Rounded up for billing
   - Init Duration: Cold start time (first run)
3. AWS Console: ${module.lambda_processor.function_url}
EOT
}

# REMOVED: SQS-related outputs
# REMOVED: ECS processor outputs
# REMOVED: Scaling instructions for ECS workers