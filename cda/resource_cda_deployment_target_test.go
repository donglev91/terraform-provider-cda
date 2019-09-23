package cda

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"net/url"
	"testing"
)

func createDeploymentTargetTemplate(name string, customType string) string {
	return fmt.Sprintf(`
resource "cda_deployment_target" "deploymentTarget" {
  name  = "%s"
  folder    = "DEFAULT"
  type  = "%s"
  dynamic_properties = {}
  custom_properties = {
    staging_remote_directory = "Remote CDA"
    staging_base_directory = "Local CDA"
  }
  description = "Deployment Target"
}	
`, name, customType)
}

func updateDeploymentTargetTemplate(name string) string {
	return fmt.Sprintf(`
resource "cda_deployment_target" "deploymentTarget" {
  name  = "%s"
  folder    = "DEFAULT"
  type  = "Generic"
  dynamic_properties = {}
  custom_properties = {}
  description = "Deployment Target"
}	
`, name)
}

func TestAccDeployment_Target(t *testing.T) {
	err := SetUpTest()
	if err != nil {
		t.Errorf("Set up test fail")
		return
	}

	resourceName := "cda_deployment_target.deploymentTarget"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testDeploymentDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: createDeploymentTargetTemplate("Testing_Name_1", "Generic"),
				Check: resource.ComposeTestCheckFunc(
					testDeploymentTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, "Testing_Name_1"),
					resource.TestCheckResourceAttr(resourceName, folder, "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, customType, "Generic"),
					resource.TestCheckResourceAttr(resourceName, description, "Deployment Target"),
					resource.TestCheckResourceAttr(resourceName, "custom_properties.staging_base_directory", "Local CDA"),
					resource.TestCheckResourceAttr(resourceName, "custom_properties.staging_remote_directory", "Remote CDA"),
				),
			},
			{
				Config: updateDeploymentTargetTemplate("Testing_Name_2"),
				Check: resource.ComposeTestCheckFunc(
					testDeploymentTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, "Testing_Name_2"),
				),
			},
		},
	})
}

func TestAccDeployment_Target_Change_Type(t *testing.T) {
	err := SetUpTest()
	if err != nil {
		t.Errorf("Set up test fail")
		return
	}

	resourceName := "cda_deployment_target.deploymentTarget"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testDeploymentDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: createDeploymentTargetTemplate("Testing_Name_1", "Generic"),
				Check: resource.ComposeTestCheckFunc(
					testDeploymentTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, "Testing_Name_1"),
					resource.TestCheckResourceAttr(resourceName, customType, "Generic"),
				),
			},
			{
				Config: createDeploymentTargetTemplate("Testing_Name_2", "Tomcat"),
				Check: resource.ComposeTestCheckFunc(
					testDeploymentTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, "Testing_Name_2"),
					resource.TestCheckResourceAttr(resourceName, customType, "Tomcat"),
				),
			},
		},
	})
}

func testDeploymentTargetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Deployment Target is set")
		}

		deploymentTargetId := rs.Primary.Attributes["id"]

		config := testAccProvider.Meta().(*Config)

		_, err := config.GetRequest("deployment_targets/" + deploymentTargetId)
		if err != nil {
			log.Printf("[ERROR] POST Request failed")
			return err
		}

		return nil
	}
}

func testDeploymentDestroy(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		deploymentTargetName := rs.Primary.Attributes[name]

		config := testAccProvider.Meta().(*Config)

		_, err := config.GetRequest("deployment_targets?name=" + url.QueryEscape(deploymentTargetName))
		if err != nil {
			log.Printf("[ERROR] POST Request failed")
			return err
		}

		return nil
	}
}
