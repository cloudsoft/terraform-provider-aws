package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"log"
	"regexp"
	"time"
)

/*

This resource is intended to be used as a base for the following resources.

It can be used as-is where the parameters are supplied as a JSON document.

Use of the wrapper resources, aws_servicecatalog_launch_role_constraint, etc, are suggested.

*/
func resourceAwsServiceCatalogConstraint() *schema.Resource {
	var awsResourceIdPattern = regexp.MustCompile("^[a-zA-Z0-9_\\-]*")
	return &schema.Resource{
		Create: resourceAwsServiceCatalogConstraintCreate,
		Read:   resourceAwsServiceCatalogConstraintRead,
		Update: resourceAwsServiceCatalogConstraintUpdate,
		Delete: resourceAwsServiceCatalogConstraintDelete,
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"parameters": {
				Type:     schema.TypeString,
				Required: true,
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					awsResourceIdPattern,
					"invalid id format"),
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					awsResourceIdPattern,
					"invalid id format"),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"LAUNCH",
					"NOTIFICATION",
					"RESOURCE_UPDATE",
					"STACKSET",
					"TEMPLATE"},
					false),
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsServiceCatalogConstraintCreate(d *schema.ResourceData, meta interface{}) error {
	jsonParameters := d.Get("parameters").(string)
	constraintType := d.Get("type").(string)
	return resourceAwsServiceCatalogConstraintCreateFromJson(d, meta, jsonParameters, constraintType)
}

func resourceAwsServiceCatalogConstraintCreateFromJson(d *schema.ResourceData, meta interface{}, jsonParameters string, constraintType string) error {
	input := servicecatalog.CreateConstraintInput{
		Parameters:  aws.String(jsonParameters),
		PortfolioId: aws.String(d.Get("portfolio_id").(string)),
		ProductId:   aws.String(d.Get("product_id").(string)),
		Type:        aws.String(constraintType),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	conn := meta.(*AWSClient).scconn
	retryCountdown := 10
	retrySleepDuration := 10 * time.Second
	for retryCountdown > 0 {
		retryCountdown--
		result, err := conn.CreateConstraint(&input)
		if err != nil {
			if scErr, ok := err.(awserr.Error); ok && scErr.Code() == servicecatalog.ErrCodeResourceNotFoundException {
				log.Printf("Resource not found - retrying...")
				time.Sleep(retrySleepDuration)
			} else {
				return fmt.Errorf("creating Constraint failed: %s", err.Error())
			}
		} else {
			d.SetId(*result.ConstraintDetail.ConstraintId)
			return resourceAwsServiceCatalogConstraintRead(d, meta)
		}
	}
	return fmt.Errorf("creating Constraint failed - retry time out")
}

func resourceAwsServiceCatalogConstraintRead(d *schema.ResourceData, meta interface{}) error {
	_, err := resourceAwsServiceCatalogConstraintReadBase(d, meta)
	if err != nil {
		return err
	}
	return nil
}

func resourceAwsServiceCatalogConstraintReadBase(d *schema.ResourceData, meta interface{}) (*servicecatalog.DescribeConstraintOutput, error) {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.DescribeConstraintInput{
		Id: aws.String(d.Id()),
	}
	constraint, err := conn.DescribeConstraint(&input)
	if err != nil {
		if scErr, ok := err.(awserr.Error); ok && scErr.Code() == servicecatalog.ErrCodeResourceNotFoundException {
			log.Printf("[WARN] Service Catalog Constraint %s not found, removing from state", d.Id())
			d.SetId("")
			return nil, nil
		}
		return nil, fmt.Errorf("reading Service Catalog Constraint '%s' failed: %s", *input.Id, err.Error())
	}
	details := constraint.ConstraintDetail
	d.Set("description", details.Description)
	d.Set("portfolio_id", details.PortfolioId)
	d.Set("product_id", details.ProductId)
	d.Set("type", details.Type)
	d.Set("owner", details.Owner)
	d.Set("status", constraint.Status)
	d.Set("parameters", constraint.ConstraintParameters)
	return constraint, nil
}

func resourceAwsServiceCatalogConstraintUpdate(d *schema.ResourceData, meta interface{}) error {
	input := servicecatalog.UpdateConstraintInput{
		Id: aws.String(d.Id()),
	}
	if d.HasChange("parameters") {
		v, _ := d.GetOk("parameters")
		input.Parameters = aws.String(v.(string))
	}
	err2 := resourceAwsServiceCatalogConstraintUpdateBase(d, meta, input)
	if err2 != nil {
		return err2
	}
	return resourceAwsServiceCatalogConstraintRead(d, meta)
}

func resourceAwsServiceCatalogConstraintUpdateBase(d *schema.ResourceData, meta interface{}, input servicecatalog.UpdateConstraintInput) error {
	input.Id = aws.String(d.Id())
	if d.HasChange("description") {
		v, _ := d.GetOk("description")
		input.Description = aws.String(v.(string))
	}
	if d.HasChange("portfolio_id") {
		//TODO - can't update - requires replace
	}
	if d.HasChange("") {
		//TODO - can't update - requires replace
	}
	if d.HasChange("") {
		//TODO - can't update - requires replace
	}
	conn := meta.(*AWSClient).scconn
	_, err := conn.UpdateConstraint(&input)
	if err != nil {
		return fmt.Errorf("updating Service Catalog Constraint '%s' failed: %s", *input.Id, err.Error())
	}
	return nil
}

func resourceAwsServiceCatalogConstraintDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	constraintId := d.Id()
	fmt.Printf("Deleting constraint: %s\n", constraintId)
	input := servicecatalog.DeleteConstraintInput{Id: aws.String(constraintId)}
	_, err := conn.DeleteConstraint(&input)
	if err != nil {
		return fmt.Errorf("deleting Service Catalog Constraint '%s' failed: %s", *input.Id, err.Error())
	}
	fmt.Printf("Deleted constraint: %s\n", constraintId)
	return nil
}
