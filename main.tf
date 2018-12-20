terraform {
  backend "s3" {
    bucket = "terraform-states.tobyjsullivan.com"
    key    = "states/chalk/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = "ap-southeast-2"
}

provider "random" {}

variable "lambda_package" {
  default = "./build/executor_lambda.zip"
}

data "aws_region" "current" {}

resource "random_id" "handler_id" {
  byte_length = 8
}

resource "aws_iam_role" "lambda_role" {
  name_prefix = "chalk-handler"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "lambda_logs" {
  name_prefix = "chalk-handler-logging"
  role        = "${aws_iam_role.lambda_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams",
        "logs:PutLogEvents",
        "logs:GetLogEvents",
        "logs:FilterLogEvents"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "executor" {
  filename                       = "${var.lambda_package}"
  source_code_hash               = "${base64sha256(file(var.lambda_package))}"
  function_name                  = "chalk-executor-${random_id.handler_id.hex}"
  handler                        = "executor"
  timeout                        = 30
  runtime                        = "go1.x"
  role                           = "${aws_iam_role.lambda_role.arn}"
}

output "executor_function_name" {
  value = "${aws_lambda_function.executor.function_name}"
}
