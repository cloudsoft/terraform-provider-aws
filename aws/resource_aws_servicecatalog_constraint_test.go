package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
	"time"
)

func TestAccAwsServiceCatalogConstraint_Basic(t *testing.T) {
	resourceName := "aws_servicecatalog_constraint.test"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	var dco servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceCatalogConstraintConfigRequirements(salt),
			},
			{
				PreConfig: testAccAwsServiceCatalogConstraintRolePrepPause(),
				Config:    testAccAwsServiceCatalogConstraintConfig(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConstraint(resourceName, &dco),
					resource.TestCheckResourceAttrSet(resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceName, "product_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "LAUNCH"),
					resource.TestCheckResourceAttrSet(resourceName, "parameters"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsServiceCatalogConstraintRolePrepPause() func() {
	return func() {
		time.Sleep(11 * time.Second)
	}
}

func testAccCheckConstraint(resourceName string, dco *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("constraint not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		input := &servicecatalog.DescribeConstraintInput{
			Id: aws.String(rs.Primary.ID),
		}
		conn := testAccProvider.Meta().(*AWSClient).scconn
		resp, err := conn.DescribeConstraint(input)
		if err != nil {
			return err
		}
		*dco = *resp
		return nil
	}
}

func testAccAwsServiceCatalogConstraintConfig(salt string) string {
	return composeConfig(
		testAccAwsServiceCatalogConstraintConfigRequirements(salt),
		`
resource "aws_servicecatalog_constraint" "test" {
  description = "description"
  parameters = <<EOF
{
  "RoleArn" : "${aws_iam_role.test.arn}"
}
EOF
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id = aws_servicecatalog_product.test.id
  type = "LAUNCH"
}
`)
}

func testAccAwsServiceCatalogConstraintConfigRequirements(salt string) string {
	return composeConfig(
		testAccAwsServiceCatalogConstraintConfig_role(salt),
		testAccAwsServiceCatalogConstraintConfig_portfolio(salt),
		testAccAwsServiceCatalogConstraintConfig_product(salt),
		testAccAwsServiceCatalogConstraintConfig_portfolioProductAssociation(),
	)
}

func testAccAwsServiceCatalogConstraintConfig_portfolioProductAssociation() string {
	return `
resource "aws_servicecatalog_portfolio_product_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    product_id = aws_servicecatalog_product.test.id
}`
}

func testAccAwsServiceCatalogConstraintConfig_product(salt string) string {
	return fmt.Sprintf(`
data "aws_region" "current" { }

resource "aws_s3_bucket" "test" {
  bucket        = "terraform-test-%[1]s"
  region        = data.aws_region.current.name
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = "test_templates_for_terraform_sc_dev1.json"
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
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = %[1]q
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = %[1]q
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.key}"
    }
  }
}`, salt)
}

func testAccAwsServiceCatalogConstraintConfig_portfolio(salt string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = "test-2"
  provider_name = "test-3"
}
`, salt)
}

func testAccAwsServiceCatalogConstraintConfig_role(salt string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "servicecatalog.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
  description = %[1]q
  path = "/testpath/"
  force_detach_policies = false
  max_session_duration = 3600
}
`, salt)
}

func testAccCheckServiceCatalogConstraintDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_constraint" {
			continue // not our monkey
		}
		input := servicecatalog.DescribeConstraintInput{Id: aws.String(rs.Primary.ID)}
		_, err := conn.DescribeConstraint(&input)
		if err == nil {
			return fmt.Errorf("constraint still exists")
		}
	}
	return nil
}
