package aws

import (
	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"fmt"
	"regexp"
	"strings"
	"testing"
)

type associationDetails struct {
	PortfolioId string
	ProductId   string
}

func TestAccAWSServiceCatalogProductAssociation_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsServiceCatalogProductAssociationResourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_servicecatalog_product_association.test", "product_id", regexp.MustCompile("^prod-.*")),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProductAssociation_product_association_disappears(t *testing.T) {
	var ad associationDetails

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatlaogProductAssociationDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsServiceCatalogProductAssociationResourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductAssociation("aws_servicecatalog_product_association.test", &ad),
					testAccCheckServiceCatlaogProductAssociationDisappears(&ad),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceCatlaogProductAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_product_association" {
			continue
		}
		x := strings.Split(rs.Primary.ID, "_")
		portfolioId := x[0]
		productId := x[1]
		input := servicecatalog.ListPortfoliosForProductInput{ProductId: aws.String(productId)}
		resp, err := conn.ListPortfoliosForProduct(&input)
		if err != nil {
			return err
		}

		for _, portfolioDetail := range resp.PortfolioDetails {
			if *portfolioDetail.Id == portfolioId {
				return fmt.Errorf("Product association still exists")
			}
		}
	}

	return nil
}

func testAccCheckProductAssociation(pr string, ad *associationDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		rs, _ := s.RootModule().Resources[pr]
		x := strings.Split(rs.Primary.ID, "_")
		portfolioId := x[0]
		productId := x[1]
		input := servicecatalog.ListPortfoliosForProductInput{ProductId: aws.String(productId)}
		resp, err := conn.ListPortfoliosForProduct(&input)
		if err != nil {
			return err
		}

		for _, portfolioDetail := range resp.PortfolioDetails {
			if *portfolioDetail.Id == portfolioId {
				ad.PortfolioId = portfolioId
				ad.ProductId = productId
				return nil
			}
		}
		return fmt.Errorf("Product association does not exist")
	}
}

func testAccCheckServiceCatlaogProductAssociationDisappears(ad *associationDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		portfolioId := ad.PortfolioId
		productId := ad.ProductId
		conn := testAccProvider.Meta().(*AWSClient).scconn

		input := servicecatalog.DisassociateProductFromPortfolioInput{
			PortfolioId: aws.String(portfolioId),
			ProductId:   aws.String(productId),
		}

		_, err := conn.DisassociateProductFromPortfolio(&input)
		if err != nil {
			return err
		}

		return nil
	}
}

const testAccCheckAwsServiceCatalogProductAssociationResourceConfig_basic = `
data "aws_caller_identity" "current" {}
variable region { default = "us-west-2" }

resource "aws_s3_bucket" "bucket" {
    bucket = "deving-me-some-tf-sc-asoc-${data.aws_caller_identity.current.account_id}-${var.region}"
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

resource "aws_servicecatalog_product" "test" {
  artifact_description = "ad"
  artifact_name = "an"
  cloud_formation_template_url = "https://s3-${var.region}.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template.key}"
  description = "test"
  distributor = "disco"
  name = "test1234"
  owner = "brett"
  product_type = "CLOUD_FORMATION_TEMPLATE"
  support_description = "a test"
  support_email = "mailid@mail.com"
  support_url = "https://url/support.html"
}

resource "aws_servicecatalog_portfolio" "test" {
  name = "test-1"
  description = "test-2"
  provider_name = "test-3"
}

resource "aws_servicecatalog_product_association" "test" {
  depends_on = ["aws_servicecatalog_product.test", "aws_servicecatalog_portfolio.test"]
  portfolio_id = "${aws_servicecatalog_portfolio.test.id}"
  product_id = "${aws_servicecatalog_product.test.id}"
}
`
