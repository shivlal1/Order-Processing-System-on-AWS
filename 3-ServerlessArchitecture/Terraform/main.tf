# Wire together modules for Lambda-based order processing

module "network" {
  source            = "./modules/network"
  service_name      = var.service_name
  container_port    = var.container_port
  health_check_path = var.health_check_path
}

# ECR for Order Receiver (keep this)
module "ecr_receiver" {
  source          = "./modules/ecr"
  repository_name = "${var.ecr_repository_name}-receiver"
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# SNS Topic for order events (keep this)
module "sns" {
  source     = "./modules/sns"
  topic_name = var.sns_topic_name
}

# REMOVED: SQS module - not needed for Lambda
# REMOVED: ECR for processor - Lambda doesn't use ECR

data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# ECS Service for Order Receiver (keep this unchanged)
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

# REMOVED: ECS Service for Order Processor - replaced by Lambda

# Build Lambda deployment package
resource "null_resource" "lambda_build" {
  triggers = {
    source_hash = filemd5("../src/order-processor-lambda/main.go")
  }

  provisioner "local-exec" {
    command = <<EOT
      cd ../src/order-processor-lambda
      GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
      zip -j function.zip bootstrap
    EOT
  }
}

# Data source to read the zip file after it's built
data "local_file" "lambda_zip" {
  filename = "../src/order-processor-lambda/function.zip"
  depends_on = [null_resource.lambda_build]
}

# Lambda function for order processing
module "lambda_processor" {
  source = "./modules/lambda"

  function_name      = "${var.service_name}-processor-lambda"
  filename           = "../src/order-processor-lambda/function.zip"
  source_code_hash   = data.local_file.lambda_zip.content_base64sha256
  execution_role_arn = data.aws_iam_role.lab_role.arn
  timeout            = var.lambda_timeout
  memory_size        = var.lambda_memory
  sns_topic_arn      = module.sns.topic_arn
  log_retention_days = var.log_retention_days

  environment_variables = {
    LOG_LEVEL = "INFO"
  }

  depends_on = [null_resource.lambda_build]
}

# Build & push Order Receiver image (keep this)
resource "docker_image" "receiver" {
  name = "${module.ecr_receiver.repository_url}:latest"

  build {
    context = "../src/order-receiver"
  }
}

resource "docker_registry_image" "receiver" {
  name = docker_image.receiver.name
}

# REMOVED: Docker build for processor - Lambda doesn't need Docker