package aws

import (
	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"fmt"
	"testing"

	"bytes"
	"text/template"
)

func TestAccAWSServiceCatalogProduct_basic(t *testing.T) {
	template := template.Must(template.New("hcl").Parse(testAccCheckAwsServiceCatalogProductResourceConfig_basic_tempate))
	var template1, template2, template3 bytes.Buffer
	template.Execute(&template1, Input{"dsc1", "dst1", "nm1", "own1", "sd1", "a@b.com", "https://url/support1.html"})
	template.Execute(&template2, Input{"dsc2", "dst2", "nm2", "own2", "sd2", "c@d.com", "https://url/support2.html"})
	template.Execute(&template3, Input{"dsc2", "dst2", "ONE_SINGLE_UPDATE", "own2", "sd2", "c@d.com", "https://url/support2.html"})

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: template1.String(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_description", "ad"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_name", "an"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "description", "dsc1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "distributor", "dst1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "name", "nm1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "owner", "own1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "product_type", "CLOUD_FORMATION_TEMPLATE"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_description", "sd1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_email", "a@b.com"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_url", "https://url/support1.html"),
				),
			},
			resource.TestStep{
				Config: template2.String(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_description", "ad"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_name", "an"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "description", "dsc2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "distributor", "dst2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "name", "nm2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "owner", "own2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "product_type", "CLOUD_FORMATION_TEMPLATE"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_description", "sd2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_email", "c@d.com"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_url", "https://url/support2.html"),
				),
			},
			resource.TestStep{
				Config: template3.String(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_description", "ad"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_name", "an"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "description", "dsc2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "distributor", "dst2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "name", "ONE_SINGLE_UPDATE"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "owner", "own2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "product_type", "CLOUD_FORMATION_TEMPLATE"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_description", "sd2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_email", "c@d.com"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_url", "https://url/support2.html"),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_disappears(t *testing.T) {
	var productViewDetail servicecatalog.ProductViewDetail
	var template1 bytes.Buffer
	template := template.Must(template.New("hcl").Parse(testAccCheckAwsServiceCatalogProductResourceConfig_basic_tempate))
	template.Execute(&template1, Input{"dsc1", "dst1", "nm1", "own1", "sd1", "a@b.com", "https://url/support1.html"})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatlaogProductDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: template1.String(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProduct("aws_servicecatalog_product.test", &productViewDetail),
					testAccCheckServiceCatlaogProductDisappears(&productViewDetail),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProduct(pr string, pd *servicecatalog.ProductViewDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		input := servicecatalog.DescribeProductAsAdminInput{}
		input.Id = aws.String(rs.Primary.ID)

		resp, err := conn.DescribeProductAsAdmin(&input)
		if err != nil {
			return err
		}

		*pd = *resp.ProductViewDetail
		return nil
	}
}

func testAccCheckServiceCatlaogProductDisappears(pd *servicecatalog.ProductViewDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn

		input := servicecatalog.DeleteProductInput{}
		input.Id = pd.ProductViewSummary.ProductId

		_, err := conn.DeleteProduct(&input)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckServiceCatlaogProductDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_portfolio" {
			continue
		}
		input := servicecatalog.DescribeProductInput{}
		input.Id = aws.String(rs.Primary.ID)

		_, err := conn.DescribeProduct(&input)
		if err == nil {
			return fmt.Errorf("Product still exists")
		}
	}

	return nil
}

type Input struct {
	Description        string
	Distributor        string
	Name               string
	Owner              string
	SupportDescription string
	SupportEmail       string
	SupportUrl         string
}

const testAccCheckAwsServiceCatalogProductResourceConfig_basic_tempate = `
data "aws_caller_identity" "current" {}
variable region { default = "us-west-2" }

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

resource "aws_servicecatalog_product" "test" {
  artifact_description = "ad"
  artifact_name = "an"
  cloud_formation_template_url = "https://s3-${var.region}.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template.key}"
  description = "{{.Description}}"
  distributor = "{{.Distributor}}"
  name = "{{.Name}}"
  owner = "{{.Owner}}"
  product_type = "CLOUD_FORMATION_TEMPLATE"
  support_description = "{{.SupportDescription}}"
  support_email = "{{.SupportEmail}}"
  support_url = "{{.SupportUrl}}"
}
`
