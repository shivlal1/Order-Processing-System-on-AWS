output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs_receiver.cluster_name
}

output "receiver_service_name" {
  description = "Name of the order receiver ECS service"
  value       = module.ecs_receiver.service_name
}

output "processor_service_name" {
  description = "Name of the order processor ECS service"
  value       = module.ecs_processor.service_name
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

# Async endpoint (Phase 3 - fast)
output "test_async_endpoint" {
  description = "Asynchronous order processing (<100ms)"
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

output "sqs_queue_url" {
  description = "URL of the SQS queue"
  value       = module.sqs.queue_url
}

output "sqs_queue_name" {
  description = "Name of the SQS queue for CloudWatch monitoring"
  value       = module.sqs.queue_name
}

# Phase 5 scaling instructions
output "scaling_instructions" {
  description = "How to scale processor workers"
  value       = <<EOT
To scale processor workers:
1. Update processor_worker_count in terraform.tfvars (e.g., 5, 20, 100)
2. Run: terraform apply
3. Monitor queue depth in CloudWatch → SQS → ${module.sqs.queue_name}
EOT
}