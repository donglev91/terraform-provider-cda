package cda

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceCdaEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceEnvironmentCreate,
		Read:   resourceEnvironmentRead,
		Update: resourceEnvironmentUpdate,
		Delete: resourceEnvironmentDelete,

		Schema: map[string]*schema.Schema{
			name: {
				Type:     schema.TypeString,
				Required: true,
			},
			customType: {
				Type:     schema.TypeString,
				Required: true,
			},
			folder: {
				Type:     schema.TypeString,
				Optional: true,
			},
			deploymentTargets: {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			dynamicProperties: {
				Type:         schema.TypeMap,
				Optional:     true,
				ValidateFunc: validateMapPropertyValue,
			},
			customProperties: {
				Type:         schema.TypeMap,
				Optional:     true,
				ValidateFunc: validateMapPropertyValue,
			},
			description: {
				Type:     schema.TypeString,
				Optional: true,
			},
			owner: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func assignTargets(d *schema.ResourceData, config *Config, m interface{}, firstCreate bool) error {
	targetMapping, found := d.GetOk(deploymentTargets)
	if !found && firstCreate {
		return nil
	}

	var body map[string]interface{}
	body = make(map[string]interface{})
	body[propertiesMap[deploymentTargets]] = targetMapping

	response, err := config.PostRequest("environments/"+d.Id()+"/deployment_targets", body)

	isCreated, errorOutput, err := GetStatus(response)
	if !isCreated {
		rollBackErr := resourceEnvironmentDelete(d, m)
		if rollBackErr == nil {
			const msg = "[ERROR] While assigning targets to environment: %s"
			if errorOutput != nil {
				log.Printf(msg, errorOutput.Error+errorOutput.Details)
				return fmt.Errorf(msg, errorOutput.Error+errorOutput.Details)
			}
			if err != nil {
				return fmt.Errorf(msg, err.Error())
			}
		} else {
			const rollbackMsg = "[ERROR] While assigning targets to environment then rollback: %s"
			return fmt.Errorf(rollbackMsg, rollBackErr.Error())
		}
	}
	return nil
}

func createEnvironmentBody(d *schema.ResourceData, config *Config) (map[string]interface{}, error) {
	var body = createBody(d, []string{name, customType}, []string{description, customProperties}, true)

	body, err := handlingBodyRequest(d, body, config)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func resourceEnvironmentCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
	body, err := createEnvironmentBody(d, config)
	if err != nil {
		return err
	}

	err = createThenValidateResponse(d, m, EnvironmentType, body)
	if err != nil {
		return err
	}
	return assignTargets(d, config, m, true)
}

func resourceEnvironmentRead(d *schema.ResourceData, m interface{}) error {
	return readThenValidateResponse(d, m, EnvironmentType)
}

func resourceEnvironmentUpdate(d *schema.ResourceData, m interface{}) error {
	if d.Id() == "" {
		return fmt.Errorf("[ERROR] Environment does not exists")
	}

	config := m.(*Config)
	if _, ok := isUpdateNeed(d, customType); ok {
		err := resourceEnvironmentDelete(d, m)
		if err != nil {
			return err
		}
		return resourceEnvironmentCreate(d, m)
	}

	var body = createUpdateBody(d, []string{name, folder, customProperties, description, owner}, true)

	err := updateThenValidateResponse(d, m, EnvironmentType, body)
	if err != nil {
		return err
	}
	return assignTargets(d, config, m, false)
}

func resourceEnvironmentDelete(d *schema.ResourceData, m interface{}) error {
	_ = resourceEnvironmentRead(d, m)
	return deleteThenValidateResponse(d, m, EnvironmentType)
}
