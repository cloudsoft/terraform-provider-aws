package aws

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"regexp"
	"time"
)

func resourceAwsServiceCatalogConstraintLaunch() *schema.Resource {
	var awsResourceIdPattern = regexp.MustCompile("^[a-zA-Z0-9_\\-]*")
	return &schema.Resource{
		Create: resourceAwsServiceCatalogConstraintLaunchCreate,
		Read:   resourceAwsServiceCatalogConstraintLaunchRead,
		Update: resourceAwsServiceCatalogConstraintLaunchUpdate,
		Delete: resourceAwsServiceCatalogConstraintLaunchDelete,
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
			// one of local_role_name or role_arn but not both
			"local_role_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_arn": {
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
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameters": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsServiceCatalogConstraintLaunchCreate(d *schema.ResourceData, meta interface{}) error {
	var jsonDoc = ""
	if localRoleName, localRoleNameProvided := d.GetOk("local_role_name"); localRoleNameProvided {
		// local role name provided
		if _, roleArnProvided := d.GetOk("role_arn"); roleArnProvided {
			return fmt.Errorf("both 'local_role_name' and 'role_arn' should not be provided")
		}
		// we have localRoleName
		jsonDoc = fmt.Sprintf(`{"LocalRoleName": "%s"}`, localRoleName.(string))
	} else {
		if roleArn, roleArnProvided := d.GetOk("role_arn"); roleArnProvided {
			// we have roleArn
			jsonDoc = fmt.Sprintf(`{"RoleArn" : "%s"}`, roleArn.(string))
		} else {
			return fmt.Errorf("either 'local_role_name' or 'role_arn' should be provided")
		}
	}
	err := resourceAwsServiceCatalogConstraintCreateFromJson(d, meta, jsonDoc)
	if err != nil {
		return err
	}
	return resourceAwsServiceCatalogConstraintLaunchRead(d, meta)
}

func resourceAwsServiceCatalogConstraintLaunchRead(d *schema.ResourceData, meta interface{}) error {
	constraint, err := resourceAwsServiceCatalogConstraintReadBase(d, meta)
	if err != nil {
		return err
	}
	var jsonDoc *string = constraint.ConstraintParameters
	var bytes []byte = []byte(*jsonDoc)
	type LaunchParameters struct {
		LocalRoleName string
		RoleArn string
	}
	var launchParameters LaunchParameters
	err = json.Unmarshal(bytes, &launchParameters)
	if err != nil {
		return err
	}
	d.Set("local_role_name", launchParameters.LocalRoleName)
	d.Set("role_arn", launchParameters.RoleArn)
	return nil
}

func resourceAwsServiceCatalogConstraintLaunchUpdate(d *schema.ResourceData, meta interface{}) error {
	//TODO
	return resourceAwsServiceCatalogConstraintLaunchRead(d, meta)
}

func resourceAwsServiceCatalogConstraintLaunchDelete(d *schema.ResourceData, meta interface{}) error {
	//TODO
	return nil
}
