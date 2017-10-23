package aws

import (
	"github.com/hashicorp/terraform/helper/resource"

	"testing"
)

func TestAccAWSServiceCatalogConstrainBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsServiceCatalogConstraintResourceConfigBasic1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_constraint.test", "description", "test"),
					resource.TestCheckResourceAttr("aws_servicecatalog_constraint.test", "owner", "658335898421"),
					resource.TestCheckResourceAttr("aws_servicecatalog_constraint.test", "parameters", "{\"RoleArn\":\"arn:aws:iam::658335898421:role/test1-me-some-sc-cons\"}"),
					resource.TestCheckResourceAttr("aws_servicecatalog_constraint.test", "type", "LAUNCH"),
				),
			},
		},
	})
}

const testAccCheckAwsServiceCatalogConstraintResourceConfigBasic1 = `
data "aws_caller_identity" "current" {}
variable region { default = "us-west-2" }

resource "aws_iam_role" "test" {
  name = "test1-me-some-sc-cons"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "servicecatalog.us-west-2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  region = "${var.region}"
  acl    = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "template" {
  bucket = "${aws_s3_bucket.bucket.id}"
  key = "test_templates_for_terraform_sc_dev.json"
  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test CF teamplate for Service Catalog terraform dev",
  "Resources": {
    "Empty": {
      "Type": "AWS::CloudFormation::WaitConditionHandle"
    }
  }
}
EOF
}

resource "aws_servicecatalog_portfolio" "test" {
  name = "test"
  description = "test-2"
  provider_name = "test-3"
}

resource "aws_servicecatalog_product" "test" {
  artifact_description = "ad"
  artifact_name = "an"
  cloud_formation_template_url = "https://s3-${var.region}.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template.key}"
  description = "a product"
  distributor = "me"
  name = "cool product"
  owner = "Brett"
  product_type = "CLOUD_FORMATION_TEMPLATE"
  support_description = "hit me up on myspace"
  support_email = "test@example.com"
  support_url = "https://www.example.com"
}

resource "aws_servicecatalog_product_association" "test" {
  depends_on = ["aws_servicecatalog_product.test", "aws_servicecatalog_portfolio.test"]
  portfolio_id = "${aws_servicecatalog_portfolio.test.id}"
  product_id = "${aws_servicecatalog_product.test.id}"
}

resource "aws_servicecatalog_constraint" "test" {
  description = "test"
  parameters = "{\"RoleArn\":\"${aws_iam_role.test.arn}\"}"
  portfolio_id = "${aws_servicecatalog_portfolio.test.id}"
  product_id = "${aws_servicecatalog_product.test.id}"
  type = "LAUNCH"
}
`
