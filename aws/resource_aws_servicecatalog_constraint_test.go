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
					resource.TestCheckResourceAttr("aws_servicecatalog_constraint.test", "owner", "1234"),
					resource.TestCheckResourceAttr("aws_servicecatalog_constraint.test", "parameters", "arn"),
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
  name = "test1-me-some-sc"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
    bucket = "deving-me-some-tf-sc-${data.aws_caller_identity.current.account_id}-${var.region}"
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

resource "aws_servicecatalog_constraint" "test" {
  description = "test constraint"
  parameters = "${aws_iam_role.test.arn}"
  portfolio_id = "${aws_servicecatalog_portfolio.test.id}"
  product_id = "${aws_servicecatalog_product.test.id}"
  type = "LAUNCH"
}
`
