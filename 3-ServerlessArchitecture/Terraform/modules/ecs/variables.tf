variable "service_name" {
  type        = string
  description = "Base name for ECS resources"
}

variable "image" {
  type        = string
  description = "ECR image URI (with tag)"
}

variable "container_port" {
  type        = number
  description = "Port your app listens on (0 for non-web services)"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnets for FARGATE tasks"
}

variable "security_group_ids" {
  type        = list(string)
  description = "SGs for FARGATE tasks"
}

variable "execution_role_arn" {
  type        = string
  description = "ECS Task Execution Role ARN"
}

variable "task_role_arn" {
  type        = string
  description = "IAM Role ARN for app permissions"
}

variable "log_group_name" {
  type        = string
  description = "CloudWatch log group name"
}

variable "region" {
  type        = string
  description = "AWS region (for awslogs driver)"
}

variable "cpu" {
  type        = string
  default     = "256"
  description = "vCPU units"
}

variable "memory" {
  type        = string
  default     = "512"
  description = "Memory (MiB)"
}

# ALB target group (optional for non-web services)
variable "target_group_arn" {
  type        = string
  description = "ARN of the ALB target group (null for non-web services)"
  default     = null
}

# Auto-scaling configuration
variable "min_capacity" {
  type        = number
  description = "Minimum number of tasks"
}

variable "max_capacity" {
  type        = number
  description = "Maximum number of tasks"
}

variable "target_cpu" {
  type        = number
  description = "Target CPU utilization percentage for scaling"
}

variable "scale_cooldown" {
  type        = number
  description = "Cooldown period in seconds between scaling activities"
}

# Environment variables for containers
variable "environment_variables" {
  type        = map(string)
  description = "Environment variables to pass to the container"
  default     = {}
}