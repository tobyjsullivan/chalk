/*
 * S3 bucket for the web app.
 */

resource "aws_s3_bucket" "web" {
  bucket = "messy.codes"

  policy = <<EOF
{
  "Version":"2012-10-17",
  "Statement":[{
	"Sid":"PublicReadGetObject",
        "Effect":"Allow",
	  "Principal": "*",
      "Action":["s3:GetObject"],
      "Resource":["arn:aws:s3:::messy.codes/*"
      ]
    }
  ]
}
EOF

  website {
    index_document = "index.html"
    error_document = "index.html"
  }
}
