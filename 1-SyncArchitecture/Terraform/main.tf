# Wire together four focused modules: network, ecr, logging, ecs.

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = var.container_port

  # NEW: Pass health check path for ALB
  health_check_path = var.health_check_path
}

module "ecr" {
  source          = "./modules/ecr"
  repository_name = var.ecr_repository_name
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# Reuse an existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}:latest"
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name

  # REMOVED: ecs_count (replaced by auto-scaling)
  # ecs_count          = var.ecs_count

  region = var.aws_region

  # Pass Fargate CPU and Memory settings
  cpu    = var.fargate_cpu     # 256 CPU units (0.25 vCPU)
  memory = var.fargate_memory  # 512 MB

  # NEW: Pass ALB target group from network module
  target_group_arn = module.network.target_group_arn

  # NEW: Auto-scaling configuration
  min_capacity   = var.min_instances    # Minimum 2 tasks
  max_capacity   = var.max_instances    # Maximum 4 tasks
  target_cpu     = var.target_cpu       # Scale at 70% CPU
  scale_cooldown = var.scale_cooldown   # 300 second cooldown
}

// Build & push the Go app image into ECR
resource "docker_image" "app" {
  # Use the URL from the ecr module, and tag it "latest"
  name = "${module.ecr.repository_url}:latest"

  build {
    # relative path from terraform/ → src/
    context = "../src"
    # Dockerfile defaults to "Dockerfile" in that context
  }
}

resource "docker_registry_image" "app" {
  # this will push :latest → ECR
  name = docker_image.app.name
}