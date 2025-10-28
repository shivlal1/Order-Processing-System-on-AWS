output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the running ECS service"
  value       = module.ecs.service_name
}

# ALB endpoint to access the service
output "load_balancer_url" {
  description = "URL of the Application Load Balancer"
  value       = "http://${module.network.alb_dns_name}"
}

# Instructions for testing synchronous endpoint
output "test_sync_endpoint" {
  description = "Example curl command to test the synchronous order processing"
  value       = "curl -X POST http://${module.network.alb_dns_name}/orders/sync -H 'Content-Type: application/json' -d '{\"customer_id\": 123, \"items\": [{\"product_id\": \"prod-1\", \"quantity\": 2, \"price\": 29.99}]}'"
}

output "health_check_endpoint" {
  description = "Health check endpoint for monitoring"
  value       = "http://${module.network.alb_dns_name}/health"
}

output "stats_endpoint" {
  description = "Stats endpoint to monitor system performance"
  value       = "http://${module.network.alb_dns_name}/stats"
}