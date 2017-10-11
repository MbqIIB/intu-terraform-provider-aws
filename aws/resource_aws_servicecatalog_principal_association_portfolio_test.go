package aws

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestAccAWSServiceCatalogPrincipalAssociationPortfolio_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatlaogPrincipalAssociationPortfolioDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsServiceCatalogPrincipalAssociationPortfolioResourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_servicecatalog_principal_association_portfolio.test1", "portfolio_id", regexp.MustCompile("^port-.*")),
				),
			},
		},
	})
}

func testAccCheckServiceCatlaogPrincipalAssociationPortfolioDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_principal_association_portfolio" {
			continue
		}
		x := strings.Split(rs.Primary.ID, "_")
		portfolioId := x[0]
		dpi := servicecatalog.DescribePortfolioInput{}
		dpi.Id = aws.String(portfolioId)
		_, err := conn.DescribePortfolio(&dpi)
		if err != nil {
			if scErr, ok := err.(awserr.Error); ok && scErr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}

		arn, err := base64.URLEncoding.DecodeString(x[2])
		if err != nil {
			return err
		}
		principalArn := string(arn)

		input := servicecatalog.ListPrincipalsForPortfolioInput{PortfolioId: aws.String(portfolioId)}
		resp, err := conn.ListPrincipalsForPortfolio(&input)
		if err != nil {
			return fmt.Errorf("%v %s", input, err.Error())
		}

		for _, p := range resp.Principals {
			if *p.PrincipalARN == principalArn {
				return fmt.Errorf("Principle association to portfolio still exists")
			}
		}
	}

	return nil
}

const testAccCheckAwsServiceCatalogPrincipalAssociationPortfolioResourceConfig_basic = `
data "aws_caller_identity" "current" {}
variable region { default = "us-west-2" }

resource "aws_iam_role" "test1" {
  name = "test1-me-some-role-assoc-for-tf-sc-1"

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

resource "aws_servicecatalog_portfolio" "test1" {
  name = "test-1"
  description = "test-2"
  provider_name = "test-3"
}

resource "aws_servicecatalog_principal_association_portfolio" "test1" {
	depends_on = ["aws_servicecatalog_portfolio.test1", "aws_iam_role.test1"]
	portfolio_id = "${aws_servicecatalog_portfolio.test1.id}"
	principal_arn = "${aws_iam_role.test1.arn}"
}
`
