package cda

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCdaDeploymentTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeploymentTargetCreate,
		Read:   resourceDeploymentTargetRead,
		Update: resourceDeploymentTargetUpdate,
		Delete: resourceDeploymentTargetDelete,

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
			agent: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func createDeploymentTargetBody(d *schema.ResourceData, config *Config) (map[string]interface{}, error) {
	var body = createBody(d, []string{name, customType}, []string{description, customProperties, agent}, true)
	body, err := handlingBodyRequest(d, body, config)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func resourceDeploymentTargetCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
	body, err := createDeploymentTargetBody(d, config)
	if err != nil {
		return err
	}
	return createThenValidateResponse(d, m, DeploymentTargetType, body)
}

func resourceDeploymentTargetUpdate(d *schema.ResourceData, m interface{}) error {
	if d.Id() == "" {
		return fmt.Errorf("[ERROR] Deployment target does not exists")
	}

	if _, ok := isUpdateNeed(d, customType); ok {
		err := resourceDeploymentTargetDelete(d, m)
		if err != nil {
			return err
		}
		return resourceDeploymentTargetCreate(d, m)
	}

	var body = createUpdateBody(d, []string{name, folder, customProperties, description, owner}, true)
	return updateThenValidateResponse(d, m, DeploymentTargetType, body)
}

func resourceDeploymentTargetRead(d *schema.ResourceData, m interface{}) error {
	return readThenValidateResponse(d, m, DeploymentTargetType)
}

func resourceDeploymentTargetDelete(d *schema.ResourceData, m interface{}) error {
	_ = resourceEnvironmentRead(d, m)
	return deleteThenValidateResponse(d, m, DeploymentTargetType)
}
