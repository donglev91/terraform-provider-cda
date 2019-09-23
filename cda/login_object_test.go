package cda

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"net/url"
	"testing"
)

func createLoginObjectTemplate(name string) string {
	return fmt.Sprintf(`
resource "cda_login_object" "myLoginObject" {
  name  = "%s"
  folder    = "DEFAULT"
  description = "Login Object Resource"
  credentials = [
    {
      agent      = "Win01"
      type       = "Windows"
      username   = "Test_01"
      password   = "Test_01"
    }
  ]
}	
`, name)
}

func updateLoginObjectTemplate(name string, description string) string {
	return fmt.Sprintf(`
resource "cda_login_object" "myLoginObject" {
  name  = "%s"
  folder    = "DEFAULT"
  owner = "100/CD/CD"
  description = "%s"
  credentials = [
    {
      agent      = "Win01"
      type       = "UNIX"
      username   = "Test_02"
      password   = "Test_02"
    }
  ]
}	
`, name, description)
}

func TestAccLoginObject(t *testing.T) {
	err := SetUpTest()
	if err != nil {
		t.Errorf("Set up test fail")
		return
	}
	var testName = "test3"
	var testNameUpdate = "test10"

	resourceName := "cda_login_object.myLoginObject"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testLoginObjectDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: createLoginObjectTemplate(testName),
				Check: resource.ComposeTestCheckFunc(
					testLoginObjectExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, testName),
					resource.TestCheckResourceAttr(resourceName, folder, "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, description, "Login Object Resource"),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credentials.0.agent", "Win01"),
					resource.TestCheckResourceAttr(resourceName, "credentials.0.type", "Windows"),
					resource.TestCheckResourceAttr(resourceName, "credentials.0.username", "Test_01"),
					resource.TestCheckResourceAttr(resourceName, "credentials.0.password", "Test_01"),
				),
			},
			{
				Config: updateLoginObjectTemplate(testNameUpdate, "Login Object Resource Update"),
				Check: resource.ComposeTestCheckFunc(
					testLoginObjectExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, name, testNameUpdate),
					resource.TestCheckResourceAttr(resourceName, description, "Login Object Resource Update"),
					resource.TestCheckResourceAttr(resourceName, owner, "100/CD/CD"),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credentials.0.agent", "Win01"),
					resource.TestCheckResourceAttr(resourceName, "credentials.0.type", "UNIX"),
					resource.TestCheckResourceAttr(resourceName, "credentials.0.username", "Test_02"),
					resource.TestCheckResourceAttr(resourceName, "credentials.0.password", "Test_02"),
				),
			},
		},
	})
}

func testLoginObjectDestroy(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		loginObjectName := rs.Primary.Attributes[name]

		config := testAccProvider.Meta().(*Config)

		_, err := config.GetRequest("logins?name=" + url.QueryEscape(loginObjectName))
		if err != nil {
			log.Printf("[ERROR] POST Request failed")
			return err
		}

		return nil
	}
}

func testLoginObjectExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Login Object is set")
		}

		loginObjectId := rs.Primary.Attributes["id"]

		config := testAccProvider.Meta().(*Config)

		_, err := config.GetRequest("logins/" + loginObjectId)
		if err != nil {
			log.Printf("[ERROR] POST Request failed")
			return err
		}

		return nil
	}
}
