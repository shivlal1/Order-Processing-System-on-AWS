resource "aws_sqs_queue" "this" {
  name = var.queue_name

  # Queue configuration
  visibility_timeout_seconds = var.visibility_timeout
  message_retention_seconds  = var.message_retention
  receive_wait_time_seconds  = var.receive_wait_time

  # Enable dead letter queue if specified
  redrive_policy = var.enable_dlq ? jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq[0].arn
    maxReceiveCount     = var.max_receive_count
  }) : null

  tags = {
    Name = var.queue_name
  }
}

# Dead Letter Queue (optional)
resource "aws_sqs_queue" "dlq" {
  count = var.enable_dlq ? 1 : 0

  name = "${var.queue_name}-dlq"

  tags = {
    Name = "${var.queue_name}-dlq"
  }
}

# Subscribe SQS queue to SNS topic if topic ARN is provided
resource "aws_sns_topic_subscription" "sqs_subscription" {
  topic_arn = var.sns_topic_arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.this.arn

  # Ensure raw message delivery is disabled so we get SNS metadata
  raw_message_delivery = false
}

# SQS Queue Policy to allow SNS to send messages
resource "aws_sqs_queue_policy" "this" {
  queue_url = aws_sqs_queue.this.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action = "sqs:SendMessage"
        Resource = aws_sqs_queue.this.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = var.sns_topic_arn
          }
        }
      }
    ]
  })
}