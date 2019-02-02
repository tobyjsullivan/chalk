
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
  handler                        = "api"
  timeout                        = 30
  runtime                        = "go1.x"
  role                           = "${aws_iam_role.lambda_role.arn}"

  environment {
    variables {
      VARIABLES_SVC = "localhost:8080" // TODO
    }
  }
}

resource "aws_api_gateway_rest_api" "api" {
  depends_on = ["aws_lambda_function.executor"]
  name       = "chalk-api-${random_id.handler_id.hex}"
}

// X /execute
resource "aws_api_gateway_resource" "execute" {
  rest_api_id = "${aws_api_gateway_rest_api.api.id}"
  parent_id   = "${aws_api_gateway_rest_api.api.root_resource_id}"
  path_part   = "execute"
}

// X /variables
resource "aws_api_gateway_resource" "variables" {
  rest_api_id = "${aws_api_gateway_rest_api.api.id}"
  parent_id   = "${aws_api_gateway_rest_api.api.root_resource_id}"
  path_part   = "variables"
}

// POST /execute
resource "aws_api_gateway_method" "post_execute" {
  rest_api_id   = "${aws_api_gateway_rest_api.api.id}"
  resource_id   = "${aws_api_gateway_resource.execute.id}"
  http_method   = "POST"
  authorization = "NONE"
}

// POST /variables
resource "aws_api_gateway_method" "post_variables" {
  rest_api_id   = "${aws_api_gateway_rest_api.api.id}"
  resource_id   = "${aws_api_gateway_resource.variables.id}"
  http_method   = "POST"
  authorization = "NONE"
}

// OPTIONS /execute
resource "aws_api_gateway_method" "options_execute" {
  rest_api_id   = "${aws_api_gateway_rest_api.api.id}"
  resource_id   = "${aws_api_gateway_resource.execute.id}"
  http_method   = "OPTIONS"
  authorization = "NONE"
}

// OPTIONS /variables
resource "aws_api_gateway_method" "options_variables" {
  rest_api_id   = "${aws_api_gateway_rest_api.api.id}"
  resource_id   = "${aws_api_gateway_resource.variables.id}"
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_post_execute" {
  rest_api_id = "${aws_api_gateway_rest_api.api.id}"
  resource_id = "${aws_api_gateway_method.post_execute.resource_id}"
  http_method = "${aws_api_gateway_method.post_execute.http_method}"

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${aws_lambda_function.executor.invoke_arn}"
}

resource "aws_api_gateway_integration" "lambda_options_execute" {
  rest_api_id = "${aws_api_gateway_rest_api.api.id}"
  resource_id = "${aws_api_gateway_method.options_execute.resource_id}"
  http_method = "${aws_api_gateway_method.options_execute.http_method}"

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${aws_lambda_function.executor.invoke_arn}"
}

resource "aws_api_gateway_integration" "lambda_post_variables" {
  rest_api_id = "${aws_api_gateway_rest_api.api.id}"
  resource_id = "${aws_api_gateway_method.post_variables.resource_id}"
  http_method = "${aws_api_gateway_method.post_variables.http_method}"

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${aws_lambda_function.executor.invoke_arn}"
}

resource "aws_api_gateway_integration" "lambda_options_variables" {
  rest_api_id = "${aws_api_gateway_rest_api.api.id}"
  resource_id = "${aws_api_gateway_method.options_variables.resource_id}"
  http_method = "${aws_api_gateway_method.options_variables.http_method}"

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${aws_lambda_function.executor.invoke_arn}"
}

resource "aws_api_gateway_deployment" "api_deployment" {
  depends_on = [
    "aws_api_gateway_integration.lambda_post_execute",
    "aws_api_gateway_integration.lambda_options_execute",
    "aws_api_gateway_integration.lambda_post_variables",
    "aws_api_gateway_integration.lambda_options_variables",
  ]

  rest_api_id = "${aws_api_gateway_rest_api.api.id}"
  stage_name  = "${var.env}"

  variables {
    api_version = "${var.api_schema_version}"
  }
}

resource "aws_lambda_permission" "apigw" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.executor.arn}"
  principal     = "apigateway.amazonaws.com"

  # The /*/* portion grants access from any method on any resource
  # within the API Gateway "REST API".
  source_arn = "${aws_api_gateway_deployment.api_deployment.execution_arn}/*/*"
}

resource "aws_api_gateway_method_settings" "api_settings" {
  depends_on = [
    "aws_api_gateway_account.account",
    "aws_api_gateway_deployment.api_deployment",
  ]

  rest_api_id = "${aws_api_gateway_rest_api.api.id}"
  stage_name  = "${var.env}"
  method_path = "*/*"

  settings {
    metrics_enabled = true
    logging_level   = "INFO"
  }
}

resource "aws_api_gateway_account" "account" {
  cloudwatch_role_arn = "${aws_iam_role.cloudwatch.arn}"
}

resource "aws_iam_role" "cloudwatch" {
  name_prefix = "chalk_api_gateway_cw_global"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "cloudwatch" {
  role = "${aws_iam_role.cloudwatch.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams",
        "logs:PutLogEvents",
        "logs:GetLogEvents",
        "logs:FilterLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
