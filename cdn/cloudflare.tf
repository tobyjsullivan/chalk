variable "cloudflare_email" {}

variable "cloudflare_token" {}

variable "api_dns" {}

variable "web_dns" {}

variable "cloudflare_zone" {
  default = "messy.codes"
}

provider "cloudflare" {
  email = "${var.cloudflare_email}"
  token = "${var.cloudflare_token}"
}

resource "cloudflare_record" "api" {
  domain = "${var.cloudflare_zone}"
  name = "api.messy.codes"
  type = "CNAME"
  value = "${var.api_dns}"
  proxied = true
}

resource "cloudflare_record" "web" {
  domain = "${var.cloudflare_zone}"
  name = "messy.codes"
  type = "CNAME"
  value = "${var.web_dns}"
  proxied = true
}