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
  default = "../build/executor_lambda.zip"
}

// Some changes require a stage redeployment. That can be invoked by updating this version.
variable "api_schema_version" {
  default = "4"
}

variable "env" {
  default = "alpha"
}

data "aws_region" "current" {}

output "aws_region" {
  value = "${data.aws_region.current.id}"
}

output "api_invoke_url" {
  value = "${aws_api_gateway_deployment.api_deployment.invoke_url}"
}

output "executor_function_name" {
  value = "${aws_lambda_function.executor.function_name}"
}

output "repo_monolith_svc_url" {
  value = "${aws_ecr_repository.monolith_svc.repository_url}"
}

output "repo_resolver_svc_url" {
  value = "${aws_ecr_repository.resolver_svc.repository_url}"
}

output "repo_api_url" {
  value = "${aws_ecr_repository.api.repository_url}"
}

output "repo_web_url" {
  value = "${aws_ecr_repository.web.repository_url}"
}

output "api_alb_dns_name" {
  value = "${aws_alb.api_alb.dns_name}"
}

output "web_alb_dns_name" {
  value = "${aws_alb.web_alb.dns_name}"
}

output "website_endpoint" {
  value = "${aws_s3_bucket.web.website_endpoint}"
}

output "ecs_cluster_arn" {
  value = "${aws_ecs_cluster.main.arn}"
}

output "api_service" {
  value = "${aws_ecs_service.api.name}"
}

output "web_service" {
  value = "${aws_ecs_service.web.name}"
}
