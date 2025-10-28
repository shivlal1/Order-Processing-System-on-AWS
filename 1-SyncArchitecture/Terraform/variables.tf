# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-west-2"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "order-processor-sync"
}

variable "service_name" {
  type    = string
  default = "order-processor-sync"
}

variable "container_port" {
  type    = number
  default = 8080
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

# Fargate CPU units (256 = 0.25 vCPU)
variable "fargate_cpu" {
  type    = string
  default = "256"
}

# Fargate Memory in MB
variable "fargate_memory" {
  type    = string
  default = "512"
}

# Health check configuration
variable "health_check_path" {
  type    = string
  default = "/health"
}

# ===== AUTO SCALING VARIABLES =====
# Minimum number of tasks - set to 1 for Phase 1 to show bottleneck
variable "min_instances" {
  type    = number
  default = 1  # Start with 1 to demonstrate the bottleneck problem
}

# Maximum number of tasks
variable "max_instances" {
  type    = number
  default = 1  # Keep at 1 for Phase 1 to show sync limitations
}

# Target CPU percentage for scaling
variable "target_cpu" {
  type    = number
  default = 70
}

# Cooldown period in seconds
variable "scale_cooldown" {
  type    = number
  default = 300
}