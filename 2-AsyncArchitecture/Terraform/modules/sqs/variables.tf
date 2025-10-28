variable "queue_name" {
  description = "Name of the SQS queue"
  type        = string
}

variable "visibility_timeout" {
  description = "Visibility timeout in seconds"
  type        = number
  default     = 30
}

variable "message_retention" {
  description = "Message retention period in seconds (4 days default)"
  type        = number
  default     = 345600
}

variable "receive_wait_time" {
  description = "Long polling wait time in seconds"
  type        = number
  default     = 20
}

variable "sns_topic_arn" {
  description = "ARN of SNS topic to subscribe to"
  type        = string
}

variable "enable_dlq" {
  description = "Enable dead letter queue"
  type        = bool
  default     = false
}

variable "max_receive_count" {
  description = "Max receive count before sending to DLQ"
  type        = number
  default     = 3
}