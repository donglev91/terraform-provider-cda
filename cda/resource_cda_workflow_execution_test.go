package cda

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"testing"
)

func TestAccWorkflowExecution_Basic(t *testing.T) {
	err := SetUpTest()
	if err != nil {
		t.Errorf("Set up test fail")
		return
	}

	resourceName := "cda_workflow_execution.basicExecution"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testWorkflowExecutionDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckExecutionConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, application, "DemoApp"),
					resource.TestCheckResourceAttr(resourceName, workflow, "deploy"),
					resource.TestCheckResourceAttr(resourceName, pack, "1"),
					resource.TestCheckResourceAttr(resourceName, deploymentProfile, "Local"),
					resource.TestCheckResourceAttr(resourceName, triggers, "true"),
				),
			},
			{
				Config: testAccCheckExecutionConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, application, "DemoApp"),
					resource.TestCheckResourceAttr(resourceName, workflow, "deploy"),
					resource.TestCheckResourceAttr(resourceName, pack, "1"),
					resource.TestCheckResourceAttr(resourceName, deploymentProfile, "Local"),
					resource.TestCheckResourceAttr(resourceName, manualApproval, "true"),
					resource.TestCheckResourceAttr(resourceName, overrideExistingComponents, "true"),
					resource.TestCheckResourceAttr(resourceName, approver, "100/CD/CD"),
					resource.TestCheckResourceAttr(resourceName, schedule, "2029-08-30T09:08:00.894Z"),
					resource.TestCheckResourceAttr(resourceName, triggers, "true"),
				),
			},
		},
	})
}

func testWorkflowExecutionDestroy(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Execution is set")
		}

		executionId := rs.Primary.Attributes["id"]

		config := testAccProvider.Meta().(*Config)

		_, err := config.GetRequest("executions/" + executionId)
		if err != nil {
			log.Printf("[ERROR] POST Request failed")
			return err
		}

		return nil
	}
}

func testAccCheckExecutionConfigUpdate() string {
	return `  
  resource "cda_workflow_execution" "basicExecution" {  
  application  = "DemoApp"
  workflow    = "deploy"
  package  = "1"
  deployment_profile = "Local" 
  override_existing_components = true
  manual_approval = true
  approver = "100/CD/CD"
  schedule = "2029-08-30T09:08:00.894Z"
  triggers = "true"
}	
`
}

func testAccCheckExecutionConfigBasic() string {
	return `  
  resource "cda_workflow_execution" "basicExecution" {  
  application  = "DemoApp"
  workflow    = "deploy"
  package  = "1"
  deployment_profile = "Local"
  triggers = "true"	
  override_existing_components = true
}	
`
}
