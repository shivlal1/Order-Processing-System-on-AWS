# Wire together modules for async order processing

module "network" {
  source            = "./modules/network"
  service_name      = var.service_name
  container_port    = var.container_port
  health_check_path = var.health_check_path
}

# ECR for Order Receiver
module "ecr_receiver" {
  source          = "./modules/ecr"
  repository_name = "${var.ecr_repository_name}-receiver"
}

# ECR for Order Processor
module "ecr_processor" {
  source          = "./modules/ecr"
  repository_name = "${var.ecr_repository_name}-processor"
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# SNS Topic for order events
module "sns" {
  source     = "./modules/sns"
  topic_name = var.sns_topic_name
}

# SQS Queue for order processing
module "sqs" {
  source = "./modules/sqs"

  queue_name         = var.sqs_queue_name
  visibility_timeout = var.sqs_visibility_timeout
  receive_wait_time  = var.sqs_receive_wait_time
  sns_topic_arn      = module.sns.topic_arn
  enable_dlq         = true
  max_receive_count  = 3
}

data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# ECS Service for Order Receiver (API)
module "ecs_receiver" {
  source = "./modules/ecs"

  service_name       = "${var.service_name}-receiver"
  image              = "${module.ecr_receiver.repository_url}:latest"
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  region             = var.aws_region
  cpu                = var.fargate_cpu
  memory             = var.fargate_memory
  target_group_arn   = module.network.target_group_arn

  # Auto-scaling for receiver
  min_capacity   = var.receiver_min_instances
  max_capacity   = var.receiver_max_instances
  target_cpu     = var.target_cpu
  scale_cooldown = var.scale_cooldown

  # Environment variables for receiver
  environment_variables = {
    SNS_TOPIC_ARN = module.sns.topic_arn
  }
}

# ECS Service for Order Processor (Background Worker)
module "ecs_processor" {
  source = "./modules/ecs"

  service_name       = "${var.service_name}-processor"
  image              = "${module.ecr_processor.repository_url}:latest"
  container_port     = 0  # Processor doesn't expose any ports
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  region             = var.aws_region
  cpu                = var.fargate_cpu
  memory             = var.fargate_memory

  # No target group for processor (not behind ALB)
  target_group_arn = null

  # Fixed count for processor (no auto-scaling initially)
  min_capacity   = var.processor_task_count
  max_capacity   = var.processor_task_count
  target_cpu     = 70
  scale_cooldown = 300

  # Environment variables for processor
  environment_variables = {
    SQS_QUEUE_URL = module.sqs.queue_url
    WORKER_COUNT  = tostring(var.processor_worker_count)
  }
}

# Build & push Order Receiver image
resource "docker_image" "receiver" {
  name = "${module.ecr_receiver.repository_url}:latest"

  build {
    context = "../src/order-receiver"
  }
}

resource "docker_registry_image" "receiver" {
  name = docker_image.receiver.name
}

# Build & push Order Processor image
resource "docker_image" "processor" {
  name = "${module.ecr_processor.repository_url}:latest"

  build {
    context = "../src/order-processor"
  }
}

resource "docker_registry_image" "processor" {
  name = docker_image.processor.name
}