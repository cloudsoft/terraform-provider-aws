package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"regexp"
	"time"
)

func resourceAwsServiceCatalogConstraint() *schema.Resource {
	var awsResourceIdPattern = regexp.MustCompile("^[a-zA-Z0-9_\\-]*")
	return &schema.Resource{
		Create: createConstraint,
		Read:   readConstraint,
		Update: updateConstraint,
		Delete: deleteConstraint,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type: schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"parameters": {
				Type: schema.TypeString,
				Required: true,
			},
			"portfolio_id": {
				Type: schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					awsResourceIdPattern,
					"invalid id format"),
			},
			"product_id": {
				Type: schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					awsResourceIdPattern,
					"invalid id format"),
			},
			"type": {
				Type: schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"LAUNCH",
					"NOTIFICATION",
					"RESOURCE_UPDATE",
					"STACKSET",
					"TEMPLATE"},
					false),
			},
		},
	}
}

func createConstraint(d *schema.ResourceData, meta interface{}) error {
	input := servicecatalog.CreateConstraintInput{
		Parameters:  aws.String(d.Get("parameters").(string)),
		PortfolioId: aws.String(d.Get("portfolio_id").(string)),
		ProductId:   aws.String(d.Get("product_id").(string)),
		Type:        aws.String(d.Get("type").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	conn := meta.(*AWSClient).scconn
	result, err := conn.CreateConstraint(&input)
	if err != nil {
		return fmt.Errorf("creating Constraint failed: %s", err.Error())
	}
	d.SetId(*result.ConstraintDetail.ConstraintId)
	return readConstraint(d, meta)
}

func readConstraint(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func updateConstraint(d *schema.ResourceData, meta interface{}) error {
	return readConstraint(d, meta)
}

func deleteConstraint(d *schema.ResourceData, meta interface{}) error {
	return readConstraint(d, meta)
}