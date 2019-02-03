data "aws_availability_zones" "available" {}

//resource "aws_vpc" "main" {
//  cidr_block = "10.10.0.0/16"
//}
//
//resource "aws_subnet" "main" {
//  count = 2
//  cidr_block = "${cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)}"
//  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
//  vpc_id = "${aws_vpc.main.id}"
//}
//
//resource "aws_security_group" "allow_all" {
//  name        = "allow_all_a_1"
//  description = "Allow all inbound traffic"
//  vpc_id      = "${aws_vpc.main.id}"
//
//  ingress {
//    protocol = "6"
//    from_port = 80
//    to_port = 8080
//    cidr_blocks = ["${aws_vpc.main.cidr_block}"]
//  }
//}

resource "aws_ecr_repository" "variables_svc" {
  name = "chalk-variables-svc"
}

resource "aws_ecr_repository" "resolver_svc" {
  name = "chalk-resolver-svc"
}

resource "aws_ecr_repository" "api" {
  name = "chalk-api"
}

//resource "aws_ecs_cluster" "main" {
//  name = "chalk-cluster"
//}
//
//resource "aws_ecs_task_definition" "chalk_api" {
//  family = "chalk-api"
//  network_mode = "awsvpc"
//  requires_compatibilities = ["FARGATE"]
//  cpu = "256"
//  memory = "512"
//
//  container_definitions = <<DEFINITION
//[
//  {
//    "cpu": 256,
//    "essential": true,
//    "image": "mongo:latest",
//    "memory": 512,
//    "name": "mongodb",
//    "networkMode": "awsvpc"
//  }
//]
//DEFINITION
//}
//
//resource "aws_ecs_service" "main" {
//  name = "tf-ecs-service-1"
//  cluster = "${aws_ecs_cluster.main.id}"
//  task_definition = "${aws_ecs_task_definition.chalk_api.arn}"
//  desired_count = 1
//  launch_type = "FARGATE"
//  network_configuration {
//    security_groups = ["${aws_security_group.allow_all.id}"]
//    subnets = ["${aws_subnet.main.*.id}"]
//  }
//}
