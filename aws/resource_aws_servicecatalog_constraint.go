package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsServiceCatalogConstraint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogConstraintCreate,
		Read:   resourceAwsServiceCatalogConstraintRead,
		Update: resourceAwsServiceCatalogConstraintUpdate,
		Delete: resourceAwsServiceCatalogConstraintDelete,

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
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: false,
				Required: true,
				ForceNew: false,
			},
			"parameters": {
				Type:     schema.TypeString,
				Computed: false,
				Required: true,
				ForceNew: true,
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Computed: false,
				Required: true,
				ForceNew: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Computed: false,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: false,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsServiceCatalogConstraintCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.CreateConstraintInput{}
	now := time.Now()
	input.IdempotencyToken = aws.String(fmt.Sprintf("%d", now.UnixNano()))

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("portfolio_id"); ok {
		input.PortfolioId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("product_id"); ok {
		input.ProductId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		input.Type = aws.String(v.(string))
	}

	for i := 0; i < 10; i++ {
		resp, err := conn.CreateConstraint(&input)
		// retry due to iam consistency errors
		// this needs to be improved to only retry on retriable errors
		// similiar to principle association
		if err != nil {
			if i < 9 {
				log.Printf("[WARN] Resource not found, retrying.")
				time.Sleep(3000 * time.Millisecond)
				continue
			}
			return fmt.Errorf("Creating ServiceCatalog constraint failed: %s (%+v) %d)", err.Error(), input, i)
		}
		d.SetId(*resp.ConstraintDetail.ConstraintId)
		break
	}
	return resourceAwsServiceCatalogConstraintRead(d, meta)
}

func resourceAwsServiceCatalogConstraintRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.DescribeConstraintInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.DescribeConstraint(&input)
	if err != nil {
		return fmt.Errorf("Reading ServiceCatalog constraint '%s' failed: %s", *input.Id, err.Error())
	}
	constraintDetail := resp.ConstraintDetail

	d.Set("description", constraintDetail.Description)
	d.Set("owner", constraintDetail.Owner)
	d.Set("parameters", resp.ConstraintParameters)
	d.Set("type", constraintDetail.Type)
	return nil
}

func resourceAwsServiceCatalogConstraintUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.UpdateConstraintInput{
		Id: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateConstraint(&input)
	if err != nil {
		return fmt.Errorf("Updating ServiceCatalog constraint '%s' failed: %s", *input.Id, err.Error())
	}
	return resourceAwsServiceCatalogConstraintRead(d, meta)
}

func resourceAwsServiceCatalogConstraintDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.DeleteConstraintInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteConstraint(&input)
	if err != nil {
		return fmt.Errorf("Deleting ServiceCatalog constraint '%s' failed: %s", *input.Id, err.Error())
	}
	return nil
}
