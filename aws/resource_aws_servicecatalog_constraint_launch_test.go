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

func TestAccAwsServiceCatalogConstraintLaunch_Basic(t *testing.T) {
	resourceName := "aws_servicecatalog_constraint_launch.test"
	roleArnResourceName := resourceName + "_a_role_arn"
	localRoleNameResourceName := resourceName + "_b_local_role_name"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	var roleArnDco servicecatalog.DescribeConstraintOutput
	var localRoleNameDco servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogConstraintLaunchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceCatalogConstraintLaunchConfigRequirements(salt),
			},
			{
				//PreConfig: testAccAwsServiceCatalogConstraintLaunchRolePrepPause(),
				Config: testAccAwsServiceCatalogConstraintLaunchConfig(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConstraintLaunch(roleArnResourceName, &roleArnDco),
					resource.TestCheckResourceAttrSet(roleArnResourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(roleArnResourceName, "product_id"),
					resource.TestCheckResourceAttr(roleArnResourceName, "description", "description"),
					resource.TestCheckResourceAttr(roleArnResourceName, "type", "LAUNCH"),
					resource.TestCheckResourceAttrSet(roleArnResourceName, "role_arn"),
					resource.TestCheckNoResourceAttr(roleArnResourceName, "local_resource_name"),

					testAccCheckConstraintLaunch(localRoleNameResourceName, &localRoleNameDco),
					resource.TestCheckResourceAttrSet(localRoleNameResourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(localRoleNameResourceName, "product_id"),
					resource.TestCheckResourceAttr(localRoleNameResourceName, "description", "description"),
					resource.TestCheckResourceAttr(localRoleNameResourceName, "type", "LAUNCH"),
					resource.TestCheckResourceAttrSet(localRoleNameResourceName, "local_role_name"),
					resource.TestCheckNoResourceAttr(roleArnResourceName, "role_arn"),
				),
			},
			{
				ResourceName:      roleArnResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      localRoleNameResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsServiceCatalogConstraintLaunchRolePrepPause() func() {
	return func() {
		time.Sleep(11 * time.Second)
	}
}

func testAccCheckConstraintLaunch(resourceName string, dco *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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

func testAccAwsServiceCatalogConstraintLaunchConfigRequirements(salt string) string {
	role := fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "tfm-test-%[1]s"
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
	portfolios := fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test_a" {
  name          = "tfm-test-%[1]s-A"
  description   = "test-2"
  provider_name = "test-3"
}
resource "aws_servicecatalog_portfolio" "test_b" {
  name          = "tfm-test-%[1]s-B"
  description   = "test-2"
  provider_name = "test-3"
}
`, salt)
	product := fmt.Sprintf(`
data "aws_region" "current" { }

resource "aws_s3_bucket" "test" {
  bucket        = "tfm-test-%[1]s"
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
  name                = "tfm-test-%[1]s"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "tfm-test-%[1]s"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.key}"
    }
  }
}`, salt)
	assocs := `
resource "aws_servicecatalog_portfolio_product_association" "test_a" {
    portfolio_id = aws_servicecatalog_portfolio.test_a.id
    product_id = aws_servicecatalog_product.test.id
}
resource "aws_servicecatalog_portfolio_product_association" "test_b" {
    portfolio_id = aws_servicecatalog_portfolio.test_b.id
    product_id = aws_servicecatalog_product.test.id
}
`
	return role + portfolios + product + assocs
}

func testAccAwsServiceCatalogConstraintLaunchConfig(salt string) string {
	requirements := testAccAwsServiceCatalogConstraintLaunchConfigRequirements(salt)
	constraint := fmt.Sprintf(`
resource "aws_servicecatalog_constraint_launch" "test_a_role_arn" {
  description = "description"
  role_arn = aws_iam_role.test.arn
  portfolio_id = aws_servicecatalog_portfolio.test_a.id
  product_id = aws_servicecatalog_product.test.id
  depends_on = [aws_servicecatalog_portfolio_product_association.test_a]
}
resource "aws_servicecatalog_constraint_launch" "test_b_local_role_name" {
  description = "description"
  local_role_name = "testpath/tfm-test-%[1]s"
  portfolio_id = aws_servicecatalog_portfolio.test_b.id
  product_id = aws_servicecatalog_product.test.id
  depends_on = [aws_servicecatalog_portfolio_product_association.test_b]
}
`,
		salt)
	template := requirements + constraint
	return template
}

func testAccCheckServiceCatalogConstraintLaunchDestroy(s *terraform.State) error {
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
