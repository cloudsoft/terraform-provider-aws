package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"time"
)

func resourceAwsServiceCatalogConstraint() *schema.Resource {
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
		Schema: map[string]*schema.Schema{},
	}
}

func createConstraint(d *schema.ResourceData, meta interface{}) error {
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