data "aws_availability_zones" "available" {}

variable "vpc_cidr" {
  default = "10.10.0.0/16"
}

variable "public_subnets_cidr" {
  type    = "list"
  default = ["10.10.0.0/24", "10.10.1.0/24"]
}

variable "availability_zones" {
  type    = "list"
  default = ["ap-southeast-2a", "ap-southeast-2b"]
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

resource "aws_ecr_repository" "web" {
  name = "chalk-web"
}

/*
 * ECS Task definitions
 */
resource "aws_ecs_task_definition" "chalk_api" {
  family                   = "chalk-api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = "${aws_iam_role.ecs_execution_role.arn}"
  task_role_arn            = "${aws_iam_role.ecs_execution_role.arn}"

  container_definitions = <<DEFINITION
[
  {
    "essential": true,
    "image": "${aws_ecr_repository.api.repository_url}",
    "name": "api",
    "environment": [
      {
        "name": "RESOLVER_SVC",
        "value": "localhost:8082"
      },
      {
        "name": "VARIABLES_SVC",
        "value": "localhost:8081"
      }
    ],
    "networkMode": "awsvpc",
    "portMappings": [
      {
        "containerPort": 8080
      }
    ],
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-group": "${aws_cloudwatch_log_group.chalk_services.name}",
        "awslogs-region": "${data.aws_region.current.name}",
        "awslogs-stream-prefix": "api"
      }
    }
  },
  {
    "essential": true,
    "image": "${aws_ecr_repository.monolith_svc.repository_url}",
    "name": "monolith-svc",
    "environment": [
      {
        "name": "PORT",
        "value": "8081"
      }
    ],
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
    "environment": [
      {
        "name": "PORT",
        "value": "8082"
      },
      {
        "name": "VARIABLES_SVC",
        "value": "localhost:8081"
      }
    ],
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

resource "aws_ecs_task_definition" "chalk_web" {
  family                   = "chalk-web"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = "${aws_iam_role.ecs_execution_role.arn}"
  task_role_arn            = "${aws_iam_role.ecs_execution_role.arn}"

  container_definitions = <<DEFINITION
[
  {
    "essential": true,
    "image": "${aws_ecr_repository.web.repository_url}",
    "name": "web",
    "environment": [
      {
        "name": "PORT",
        "value": "8080"
      }
    ],
    "networkMode": "awsvpc",
    "portMappings": [
      {
        "containerPort": 8080
      }
    ],
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-group": "${aws_cloudwatch_log_group.chalk_services.name}",
        "awslogs-region": "${data.aws_region.current.name}",
        "awslogs-stream-prefix": "web"
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
resource "aws_ecs_service" "api" {
  name            = "chalk-api-${var.env}"
  cluster         = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.chalk_api.arn}"
  desired_count   = 1
  launch_type     = "FARGATE"

  depends_on = [
    "aws_iam_role_policy.ecs_execution_role_policy",
    "aws_alb_target_group.api_alb_target_group",
  ]

  network_configuration {
    security_groups = [
      "${aws_security_group.vpc_default.id}",
      "${aws_security_group.ecs_service.id}",
    ]

    subnets = [
      "${aws_subnet.public_subnet.*.id}",
    ]

    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = "${aws_alb_target_group.api_alb_target_group.arn}"
    container_name   = "api"
    container_port   = "8080"
  }
}

resource "aws_ecs_service" "web" {
  name            = "chalk-web-${var.env}"
  cluster         = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.chalk_web.arn}"
  desired_count   = 1
  launch_type     = "FARGATE"

  depends_on = [
    "aws_iam_role_policy.ecs_execution_role_policy",
    "aws_alb_target_group.web_alb_target_group",
  ]

  network_configuration {
    security_groups = [
      "${aws_security_group.vpc_default.id}",
      "${aws_security_group.ecs_service.id}",
    ]

    subnets = [
      "${aws_subnet.public_subnet.*.id}",
    ]

    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = "${aws_alb_target_group.web_alb_target_group.arn}"
    container_name   = "web"
    container_port   = "8080"
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
 * Load balancers
 */
resource "aws_alb_target_group" "api_alb_target_group" {
  name        = "chalk-api-${var.env}"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = "${aws_vpc.main.id}"
  target_type = "ip"

  lifecycle {
    create_before_destroy = true
  }

  health_check {
    path = "/health"
  }

  depends_on = ["aws_alb.api_alb"]
}

resource "aws_alb_target_group" "web_alb_target_group" {
  name        = "chalk-web--${var.env}"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = "${aws_vpc.main.id}"
  target_type = "ip"

  lifecycle {
    create_before_destroy = true
  }

  health_check {
    path = "/health"
  }

  depends_on = ["aws_alb.web_alb"]
}

resource "aws_security_group" "alb_inbound_sg" {
  name_prefix = "chalk-inbound-sg-"
  description = "Allow HTTP from Anywhere into ALB"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8
    to_port     = 0
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_alb" "api_alb" {
  name            = "chalk-api"
  subnets         = ["${aws_subnet.public_subnet.*.id}"]
  security_groups = ["${aws_security_group.vpc_default.id}", "${aws_security_group.alb_inbound_sg.id}"]
}

resource "aws_alb_listener" "api_alb_listener" {
  load_balancer_arn = "${aws_alb.api_alb.arn}"
  port              = "80"
  protocol          = "HTTP"
  depends_on        = ["aws_alb_target_group.api_alb_target_group"]

  default_action {
    target_group_arn = "${aws_alb_target_group.api_alb_target_group.arn}"
    type             = "forward"
  }
}

resource "aws_alb" "web_alb" {
  name            = "chalk-web"
  subnets         = ["${aws_subnet.public_subnet.*.id}"]
  security_groups = ["${aws_security_group.vpc_default.id}", "${aws_security_group.alb_inbound_sg.id}"]
}

resource "aws_alb_listener" "web_alb_listener" {
  load_balancer_arn = "${aws_alb.web_alb.arn}"
  port              = "80"
  protocol          = "HTTP"
  depends_on        = ["aws_alb_target_group.web_alb_target_group"]

  default_action {
    target_group_arn = "${aws_alb_target_group.web_alb_target_group.arn}"
    type             = "forward"
  }
}


/*
 * Permissions
 */
resource "aws_security_group" "vpc_default" {
  name_prefix = "chalk-default-sg-"
  description = "Default security group to allow inbound/outbound from the VPC"
  vpc_id      = "${aws_vpc.main.id}"
  depends_on  = ["aws_vpc.main"]

  ingress {
    from_port = "0"
    to_port   = "0"
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port = "0"
    to_port   = "0"
    protocol  = "-1"
    self      = true
  }
}

resource "aws_iam_role" "ecs_execution_role" {
  name_prefix = "chalk-ecs_task_execution_role"

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
  name_prefix = "chalk-ecs_execution_role_policy"
  role        = "${aws_iam_role.ecs_execution_role.id}"

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
