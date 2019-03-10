terraform {
  backend "s3" {
    bucket = "terraform-states.tobyjsullivan.com"
    key    = "states/chalk/cdn.tfstate"
    region = "us-east-1"
  }
}
