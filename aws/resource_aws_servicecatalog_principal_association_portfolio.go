package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsServiceCatalogPrincipalAssociationPortfolio() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogPrincipalAssociationPortfolioCreate,
		Read:   resourceAwsServiceCatalogPrincipalAssociationPortfolioRead,
		Delete: resourceAwsServiceCatalogPrincipalAssociationPortfolioDelete,

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
			"principal_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsServiceCatalogPrincipalAssociationPortfolioCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.AssociatePrincipalWithPortfolioInput{PrincipalType: aws.String("IAM")}
	if v, ok := d.GetOk("portfolio_id"); ok {
		input.PortfolioId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("principal_arn"); ok {
		input.PrincipalARN = aws.String(v.(string))
	}

	id := principalAssociationPortfolio(input.PortfolioId, input.PrincipalType, input.PrincipalARN)

	log.Printf("[DEBUG] Creating Service Catalog Principle Association Portfolio: %#v", input)
	_, err := conn.AssociatePrincipalWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Adding Service Catalog Principle Association Portfolio '%s' failed: %s", id, err.Error())
	}
	d.SetId(id)

	err = waitForPrincipal(conn, *input.PortfolioId, *input.PrincipalARN)
	if err != nil {
		return err
	}
	return resourceAwsServiceCatalogPrincipalAssociationPortfolioRead(d, meta)
}

func resourceAwsServiceCatalogPrincipalAssociationPortfolioRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	id := d.Id()
	x := strings.Split(id, "_")
	portfolioId := x[0]
	principalType := x[1]

	arn, err := base64.URLEncoding.DecodeString(x[2])
	if err != nil {
		return fmt.Errorf("Reading Service Catalog Principle Association Portfolio '%s' failed: %s", portfolioId, err.Error())
	}
	principalArn := string(arn)

	resp, err := listPrincipalsForPortfolio(conn, portfolioId)
	if err != nil {
		return err
	}

	for _, p := range resp.Principals {
		if *p.PrincipalARN == principalArn {
			d.Set("portfolio_id", portfolioId)
			d.Set("principal_type", principalType)
			d.Set("principal_arn", principalArn)
			return nil
		}
	}

	return fmt.Errorf("Reading Service Catalog Principle Association Portfolio '%s' failed: 'not found'.", portfolioId)
}

func waitForPrincipal(conn *servicecatalog.ServiceCatalog, portfolioId, principalArn string) error {
	var successCount int = 0
	var timeout time.Duration = 300 * time.Second
	var requiredSuccesfulCheck = 3 // check multi times as updates to AIM are eventually consistent

	err := resource.Retry(timeout,
		func() *resource.RetryError {
			resp, err := listPrincipalsForPortfolio(conn, portfolioId)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			for _, p := range resp.Principals {
				if *p.PrincipalARN == principalArn {
					successCount = successCount + 1
					if successCount == requiredSuccesfulCheck {
						return nil
					}
					return resource.RetryableError(fmt.Errorf("Not enough successes yet. Stay tuned..."))
				}
			}

			return resource.RetryableError(fmt.Errorf("Principal not available yet..."))
		})
	return err
}

func listPrincipalsForPortfolio(conn *servicecatalog.ServiceCatalog, portfolioId string) (*servicecatalog.ListPrincipalsForPortfolioOutput, error) {
	input := servicecatalog.ListPrincipalsForPortfolioInput{PortfolioId: aws.String(portfolioId)}
	resp, err := conn.ListPrincipalsForPortfolio(&input)
	if err != nil {
		return nil, fmt.Errorf("Reading Service Catalog Principle Association Portfolio '%s' failed: %s", portfolioId, err.Error())
	}
	return resp, nil
}

func resourceAwsServiceCatalogPrincipalAssociationPortfolioDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	id := d.Id()
	x := strings.Split(id, "_")
	portfolioId := x[0]
	arn, err := base64.URLEncoding.DecodeString(x[2])
	if err != nil {
		return fmt.Errorf("Deleting Service Catalog Principle Association Portfolio '%s' failed: %s", id, err.Error())
	}
	principalArn := string(arn)

	input := servicecatalog.DisassociatePrincipalFromPortfolioInput{
		PortfolioId:  aws.String(portfolioId),
		PrincipalARN: aws.String(principalArn),
	}

	_, err = conn.DisassociatePrincipalFromPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Deleting Service Catalog Principle Association Portfolio '%s' failed: %s", id, err.Error())
	}
	return nil
}

func principalAssociationPortfolio(portfolioId, principalType, principalARN *string) string {
	encodedARN := base64.URLEncoding.EncodeToString([]byte(*principalARN))
	return *portfolioId + "_" + *principalType + "_" + encodedARN
}
