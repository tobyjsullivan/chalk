data "aws_availability_zones" "available" {}

variable "vpc_cidr" {
  default = "10.10.0.0/16"
}

variable "public_subnets_cidr" {
  type = "list"
  default = ["10.10.0.0/24"]
}

variable "private_subnets_cidr" {
  type = "list"
  default = ["10.10.1.0/24"]
}

variable "availability_zones" {
  type = "list"
  default = ["ap-southeast-2a"]
}

/*
 * Docker image repositories
 */
resource "aws_ecr_repository" "monolith_svc" {
  name = "chalk-monolith-svc"
}

resource "aws_ecr_repository" "resolver_svc" {
  name = "chalk-resolver-svc"
}

resource "aws_ecr_repository" "api" {
  name = "chalk-api"
}

/*
 * ECS Task definitions
 */
resource "aws_ecs_task_definition" "chalk_api" {
  family = "chalk-api"
  network_mode = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu = "256"
  memory = "512"
  execution_role_arn = "${aws_iam_role.ecs_execution_role.arn}"

  container_definitions = <<DEFINITION
[
  {
    "essential": true,
    "image": "${aws_ecr_repository.monolith_svc.repository_url}",
    "name": "monolith-svc",
    "networkMode": "awsvpc",
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-group": "${aws_cloudwatch_log_group.chalk_services.name}",
        "awslogs-region": "${data.aws_region.current.name}",
        "awslogs-stream-prefix": "monolith-svc"
      }
    }
  },
  {
    "essential": true,
    "image": "${aws_ecr_repository.resolver_svc.repository_url}",
    "name": "resolver-svc",
    "networkMode": "awsvpc",
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-group": "${aws_cloudwatch_log_group.chalk_services.name}",
        "awslogs-region": "${data.aws_region.current.name}",
        "awslogs-stream-prefix": "repository-svc"
      }
    }
  }
]
DEFINITION
}

/*
 * ECS Fargate Cluster
 */
resource "aws_ecs_cluster" "main" {
  name = "chalk-cluster"
}

/*
 * Chalk Service
 */
resource "aws_ecs_service" "main" {
  name = "chalk-backend-${var.env}-2"
  cluster = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.chalk_api.arn}"
  desired_count = 1
  launch_type = "FARGATE"
  depends_on = ["aws_iam_role_policy.ecs_execution_role_policy"]

  network_configuration {
    security_groups = ["${aws_security_group.ecs_service.id}"]
    subnets = ["${aws_subnet.public_subnet.*.id}"]
    assign_public_ip = true
  }
}

resource "aws_security_group" "ecs_service" {
  name        = "allow_all_a_1"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"


  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8
    to_port     = 0
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

}

/*
 * The VPC
 */
resource "aws_vpc" "main" {
  cidr_block = "${var.vpc_cidr}"

  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_subnet" "public_subnet" {
  vpc_id                  = "${aws_vpc.main.id}"
  count                   = "${length(var.public_subnets_cidr)}"
  cidr_block              = "${element(var.public_subnets_cidr, count.index)}"
  availability_zone       = "${element(var.availability_zones, count.index)}"
  map_public_ip_on_launch = true
}

resource "aws_route_table" "public" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route" "public_internet_gateway" {
  route_table_id         = "${aws_route_table.public.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.ig.id}"
}

resource "aws_route_table_association" "public" {
  count          = "${length(var.public_subnets_cidr)}"
  subnet_id      = "${element(aws_subnet.public_subnet.*.id, count.index)}"
  route_table_id = "${aws_route_table.public.id}"
}

resource "aws_internet_gateway" "ig" {
  vpc_id = "${aws_vpc.main.id}"
}

/*
 * Permissions Management
 */
resource "aws_iam_role" "ecs_execution_role" {
  name_prefix  = "chalk-ecs_task_execution_role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "ecs_execution_role_policy" {
  name_prefix   = "chalk-ecs_execution_role_policy"
  role   = "${aws_iam_role.ecs_execution_role.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

/*
 * Observability
 */
resource "aws_cloudwatch_log_group" "chalk_services" {
  name_prefix = "chalk-svcs_"

  tags {
    Environment = "${var.env}"
    Application = "Chalk"
  }
}