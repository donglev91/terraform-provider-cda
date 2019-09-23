package cda

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"net/url"
	"testing"
)

//Expect following CDA preparation before enable this test.
//  Application: DemoApp
// 	Environment: Local
//  Two deployment target: Local Tomcat, Local SQLLite DB

func TestAccDeploymentProfile_Basic(t *testing.T) {
	err := SetUpTest()
	if err != nil {
		t.Errorf("Set up test fail")
		return
	}

	resourceName := "cda_deployment_profile.basicDeploymentProfile"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testDeploymentProfileDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccCreateDeploymentProfileConfigBasic("Testing_Name_1"),
				Check: resource.ComposeTestCheckFunc(
					testDeploymentProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, "Testing_Name_1"),
					resource.TestCheckResourceAttr(resourceName, folder, "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, application, "DemoApp"),
					resource.TestCheckResourceAttr(resourceName, environment, "Local"),
					resource.TestCheckResourceAttr(resourceName, description, "Create Profile"),
					resource.TestCheckResourceAttr(resourceName, owner, "100/CD/CD"),
					resource.TestCheckResourceAttr(resourceName, loginObject, "Sample"),
				),
			},
			{
				Config: testAccCheckUpdateDeploymentProfileConfigBasic("Testing_Name_2"),
				Check: resource.ComposeTestCheckFunc(
					testDeploymentProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, "Testing_Name_2"),
					resource.TestCheckResourceAttr(resourceName, folder, "DEMOAPP"),
					resource.TestCheckResourceAttr(resourceName, application, "DemoApp"),
					resource.TestCheckResourceAttr(resourceName, environment, "Local"),
					resource.TestCheckResourceAttr(resourceName, description, "Update Profile"),
					resource.TestCheckResourceAttr(resourceName, owner, "100/CD/CD"),
					resource.TestCheckResourceAttr(resourceName, loginObject, "Sample"),
					resource.TestCheckResourceAttr(resourceName, "deployment_map.database", "Local SQLLite DB,Local Tomcat"),
					resource.TestCheckResourceAttr(resourceName, "deployment_map.webapp", "Local Tomcat"),
				),
			},
		},
	})
}

func TestAccDeploymentProfile_HasTargetMapping(t *testing.T) {
	err := SetUpTest()
	if err != nil {
		_ = fmt.Errorf("Set up test fail")
		return
	}

	resourceName := "cda_deployment_profile.basicDeploymentProfileWithMapping"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testDeploymentProfileDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCreateDeploymentProfileWithTargetMappingConfigBasic("Testing_Name_3"),
				Check: resource.ComposeTestCheckFunc(
					testDeploymentProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, "Testing_Name_3"),
					resource.TestCheckResourceAttr(resourceName, folder, "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, description, "Description with target mapping"),
				),
			},
		},
	})
}

func testDeploymentProfileExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Deployment Profile is set")
		}

		profileId := rs.Primary.Attributes["id"]

		config := testAccProvider.Meta().(*Config)

		_, err := config.GetRequest("profiles/" + profileId)
		if err != nil {
			log.Printf("[ERROR] POST Request failed")
			return err
		}

		return nil
	}
}

func testDeploymentProfileDestroy(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		profileName := rs.Primary.Attributes[name]

		config := testAccProvider.Meta().(*Config)

		response, err := config.GetRequest("profiles/?name=" + url.QueryEscape(profileName))

		if err != nil {
			log.Printf("[ERROR] POST Request failed")
			return err
		}

		listResponse, err := convertListResponse(response)

		if err != nil {
			return err
		}

		if listResponse.Total > 0 {
			return fmt.Errorf("Deployment Profile existed")
		}

		return nil
	}
}

func testAccCheckUpdateDeploymentProfileConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "cda_deployment_profile" "basicDeploymentProfile" {
  name  = "%s"
  folder    = "DEMOAPP"
  application  = "DemoApp"
  environment  = "Local"
  description = "Update Profile"
  login_object = "Sample"
  owner = "100/CD/CD"
  deployment_map = {"database" = "Local SQLLite DB,Local Tomcat", "webapp" = "Local Tomcat"}
}	
`, name)
}

func testAccCreateDeploymentProfileConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "cda_deployment_profile" "basicDeploymentProfile" {
  name  = "%s"
  folder    = "DEFAULT"
  application  = "DemoApp"
  environment  = "Local"
  description = "Create Profile"
  login_object = "Sample"
  owner = "100/CD/CD"
  deployment_map = {}
}	
`, name)
}

func testAccCheckCreateDeploymentProfileWithTargetMappingConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "cda_deployment_profile" "basicDeploymentProfileWithMapping" {
  name  = "%s"
  folder    = "DEFAULT"
  application  = "DemoApp"
  environment  = "Local"
  description = "Description with target mapping"
  deployment_map = {"database" = "Local SQLLite DB,Local Tomcat"}
}	
`, name)
}