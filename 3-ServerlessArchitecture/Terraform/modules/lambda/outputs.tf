output "function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.this.function_name
}

output "function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.this.arn
}

output "log_group_name" {
  description = "CloudWatch log group name"
  value       = aws_cloudwatch_log_group.lambda_logs.name
}

output "function_url" {
  description = "URL to view function in AWS console"
  value       = "https://console.aws.amazon.com/lambda/home?region=${aws_lambda_function.this.arn != "" ? split(":", aws_lambda_function.this.arn)[3] : "us-east-1"}#/functions/${aws_lambda_function.this.function_name}"
}