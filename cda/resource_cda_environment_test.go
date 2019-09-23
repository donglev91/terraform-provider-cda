package cda

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"net/url"
	"testing"
)

func TestAccEnvironment_Basic(t *testing.T) {
	err := SetUpTest()
	if err != nil {
		t.Errorf("Set up test fail")
		return
	}

	resourceName := "cda_environment.basicEnvironment"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testEnvironmentDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEnvironmentConfigBasic("Testing_Name_1", "Generic" ),
				Check: resource.ComposeTestCheckFunc(
					testEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, "Testing_Name_1"),
					resource.TestCheckResourceAttr(resourceName, folder, "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, customType, "Generic"),
					resource.TestCheckResourceAttr(resourceName, description, "Description"),
				),
			},
			{
				Config: testAccCheckUpdateEnvironmentConfigBasic("Testing_Name_2"),
				Check: resource.ComposeTestCheckFunc(
					testEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, "Testing_Name_2"),
					resource.TestCheckResourceAttr(resourceName, description, "Description Update"),
				),
			},
		},
	})
}

func TestAccEnvironment_ChangeType(t *testing.T) {
	err := SetUpTest()
	if err != nil {
		t.Errorf("Set up test fail")
		return
	}

	resourceName := "cda_environment.basicEnvironment"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testEnvironmentDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEnvironmentConfigBasic("Testing_Name_1", "Generic" ),
				Check: resource.ComposeTestCheckFunc(
					testEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, customType, "Generic"),
				),
			},
			{
				Config: testAccCheckEnvironmentConfigBasic("Testing_Name_2", "Automic"),
				Check: resource.ComposeTestCheckFunc(
					testEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, customType, "Automic"),
				),
			},
		},
	})
}

//Expect following CDA preparation before enable this test.
//Generic environment has:
// 	One custom property name: TestCustomProperty
//  One dynamic property name:  /TestDynamicProperty
//  Two deployment target name: Local Tomcat, Local SQLLite DB
/*
func TestAccEnvironment_HasDynamic(t *testing.T) {
	err := SetUpTest();
	if err != nil {
		fmt.Errorf("Set up test fail")
		return;
	}

	resourceName := "cda_environment.dynamicEnvironment"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testEnvironmentDestroy(resourceName),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckEnvironmentConfigDynamic("Testing_Name_3"),
				Check: resource.ComposeTestCheckFunc(
					testEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, "Testing_Name_3"),
					resource.TestCheckResourceAttr(resourceName, folder, "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, customType, "Generic"),
					resource.TestCheckResourceAttr(resourceName, description, "Description"),
					resource.TestCheckResourceAttr(resourceName, "dynamic_properties.TestDynamicProperty", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "custom_properties.TestCustomProperty", "custome"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0", "Local Tomcat"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.1", "Local SQLLite DB"),
				),
			},
		},
	})
}
*/

func testEnvironmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Environment is set")
		}

		environmentId := rs.Primary.Attributes["id"]

		config := testAccProvider.Meta().(*Config)

		_, err := config.GetRequest("environments/" + environmentId)
		if err != nil {
			log.Printf("[ERROR] POST Request failed")
			return err
		}

		return nil
	}
}

func testEnvironmentDestroy(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		environmentName := rs.Primary.Attributes[name]

		config := testAccProvider.Meta().(*Config)

		response, err := config.GetRequest("environments/?name=" + url.QueryEscape(environmentName))

		if err != nil {
			log.Printf("[ERROR] POST Request failed")
			return err
		}

		listResponse, err := convertListResponse(response)

		if err != nil {
			return err
		}

		if listResponse.Total > 0 {
			return fmt.Errorf("Environment existed")
		}

		return nil
	}
}

func testAccCheckUpdateEnvironmentConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "cda_environment" "basicEnvironment" {
  name  = "%s"
  folder    = "DEFAULT"
  type  = "Generic"
  dynamic_properties = {}
  custom_properties = {}
  deployment_targets = []
  description = "Description Update"
}	
`, name)
}

func testAccCheckEnvironmentConfigBasic(name string, customType string) string {
	return fmt.Sprintf(`
resource "cda_environment" "basicEnvironment" {
  name  = "%s"
  folder    = "DEFAULT"
  type  = "%s"
  dynamic_properties = {}
  custom_properties = {}
  deployment_targets = []
  description = "Description"
}	
`, name, customType)
}

/*
func testAccCheckEnvironmentConfigDynamic(name string) string {
	return fmt.Sprintf(`
resource "cda_environment" "dynamicEnvironment" {
  name  = "%s"
  folder = "DEFAULT"
  type  = "Generic"
  dynamic_properties = {"TestDynamicProperty" = "Value1"}
  custom_properties = {"TestCustomProperty" = "custome"}
  deployment_targets = ["Local Tomcat", "Local SQLLite DB"]
  description = "Description"
  owner = "100/AUTOMIC/AUTOMIC"
}
`, name)
} */