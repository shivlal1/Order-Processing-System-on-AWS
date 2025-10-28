resource "aws_sns_topic" "this" {
  name = var.topic_name

  tags = {
    Name = var.topic_name
  }
}