package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsServiceCatalogProductAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogProductAssociationCreate,
		Read:   resourceAwsServiceCatalogProductAssociationRead,
		Delete: resourceAwsServiceCatalogProductAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_portfolio_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsServiceCatalogProductAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.AssociateProductWithPortfolioInput{}
	if v, ok := d.GetOk("portfolio_id"); ok {
		input.PortfolioId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("product_id"); ok {
		input.ProductId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_portfolio_id"); ok {
		input.SourcePortfolioId = aws.String(v.(string))
	}

	id := productAssociationId(input.PortfolioId, input.ProductId, input.SourcePortfolioId)

	log.Printf("[DEBUG] Creating Service Catalog Product Association: %#v", input)
	_, err := conn.AssociateProductWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Adding Service Catalog Product Association '%s' failed: %s", id, err.Error())
	}
	d.SetId(id)

	return resourceAwsServiceCatalogProductAssociationRead(d, meta)
}

func resourceAwsServiceCatalogProductAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	x := strings.Split(d.Id(), "_")
	if len(x) != 3 {
		return fmt.Errorf("Reading Service Catalog Product Association '%s' failed: Error parsing id", d.Id())
	}
	portfolioId := x[0]
	productId := x[1]
	sourcePortfolioId := x[2]

	input := servicecatalog.ListPortfoliosForProductInput{ProductId: aws.String(productId)}
	resp, err := conn.ListPortfoliosForProduct(&input)
	if err != nil {
		if scErr, ok := err.(awserr.Error); ok && scErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Service Catalog Product %q not found, removing association %q from state", productId, d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading Service Catalog Product Association '%s' failed: %s", d.Id(), err.Error())
	}

	for _, portfolioDetail := range resp.PortfolioDetails {
		if *portfolioDetail.Id == portfolioId {
			d.Set("product_id", productId)
			d.Set("portfolio_id", portfolioId)
			d.Set("source_portfolio_id", sourcePortfolioId)
			return nil
		}
	}

	log.Printf("[WARN] Service Catalog Product Association %q not found, removing from state", d.Id())
	d.SetId("")
	return nil
}

func resourceAwsServiceCatalogProductAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	id := d.Id()

	x := strings.Split(id, "_")
	if len(x) != 3 {
		return fmt.Errorf("Deleting Service Catalog Product Association '%s' failed: Error parsing id", d.Id())
	}
	portfolioId := x[0]
	productId := x[1]

	input := servicecatalog.DisassociateProductFromPortfolioInput{
		PortfolioId: aws.String(portfolioId),
		ProductId:   aws.String(productId),
	}

	_, err := conn.DisassociateProductFromPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Deleting ServiceCatalog Product Association '%s' failed: %s", id, err.Error())
	}
	return nil
}

func productAssociationId(portfolioId, productId, sourcePortfolioId *string) string {
	if sourcePortfolioId == nil {
		return *portfolioId + "_" + *productId + "_"
	}

	return *portfolioId + "_" + *productId + "_" + *sourcePortfolioId
}
