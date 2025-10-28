# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-west-2"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "order-processor"
}

variable "service_name" {
  type    = string
  default = "order-processor"
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

# ===== RECEIVER SERVICE SCALING =====
variable "receiver_min_instances" {
  type    = number
  default = 2
}

variable "receiver_max_instances" {
  type    = number
  default = 4
}

# ===== LAMBDA CONFIGURATION =====
variable "lambda_memory" {
  type    = number
  default = 512  # MB
}

variable "lambda_timeout" {
  type    = number
  default = 10  # seconds (3s processing + buffer)
}

# ===== AUTO SCALING SETTINGS =====
variable "target_cpu" {
  type    = number
  default = 70
}

variable "scale_cooldown" {
  type    = number
  default = 300
}

# ===== SNS CONFIGURATION =====
variable "sns_topic_name" {
  type    = string
  default = "order-processing-events"
}

# REMOVED: All SQS-related variables
# REMOVED: processor_task_count
# REMOVED: processor_worker_count