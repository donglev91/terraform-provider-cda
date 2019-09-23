package cda

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"math"
	"math/rand"
	"strconv"
)

func resourceCdaWorkflowExecution() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkflowExecutionCreate,
		Read:   resourceWorkflowExecutionRead,
		Update: resourceWorkflowExecutionUpdate,
		Delete: resourceWorkflowExecutionDelete,

		Schema: map[string]*schema.Schema{
			application: {
				Type:     schema.TypeString,
				Required: true,
			},
			workflow: {
				Type:     schema.TypeString,
				Required: true,
			},
			pack: {
				Type:     schema.TypeString,
				Required: true,
			},
			deploymentProfile: {
				Type:     schema.TypeString,
				Required: true,
			},
			manualApproval: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			approver: {
				Type:     schema.TypeString,
				Optional: true,
			},
			schedule: {
				Type:     schema.TypeString,
				Optional: true,
			},
			overrideExistingComponents: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			monitorUrl: {
				Type:     schema.TypeString,
				Computed: true,
			},
			installationUrl: {
				Type:     schema.TypeString,
				Computed: true,
			},
			triggers: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			overridesApplication: {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ValidateFunc: validateMapPropertyValue,
				Optional:     true,
			},
			overridesPackage: {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ValidateFunc: validateMapPropertyValue,
				Optional:     true,
			},
			overridesWorkflow: {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ValidateFunc: validateMapPropertyValue,
				Optional:     true,
			},
			overridesComponent: {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeMap,
					ValidateFunc: validateMapPropertyValue,
				},
				Optional: true,
			},
		},
	}
}

func resourceWorkflowExecutionUpdate(d *schema.ResourceData, m interface{}) error {
	if isExecutionDisable(d) {
		return nil
	}

	d.SetId("")
	return resourceWorkflowExecutionCreate(d, m)
}

func isExecutionDisable(d *schema.ResourceData) bool {
	if triggerProperty, exist := d.GetOkExists(triggers); exist {
		return !triggerProperty.(bool)
	}
	return false
}

func resourceWorkflowExecutionRead(d *schema.ResourceData, m interface{}) error {
	if isExecutionDisable(d) {
		return nil
	}

	id := d.Id()
	if id != "" {
		config := m.(*Config)
		response, _ := config.GetRequest("executions/" + id)
		executionResponse, err := convertExecutionResponse(response)
		if err == nil {
			_ = d.Set("installation_url", executionResponse.InstallationUrl)
			_ = d.Set("monitor_url", executionResponse.MonitorUrl)
		}
	}
	return nil
}

func cloneMap(m map[string]interface{}, ignoreKey string) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		if k != ignoreKey {
			result[k] = v
		}
	}
	return result
}

func resourceWorkflowExecutionCreate(d *schema.ResourceData, m interface{}) error {
	if isExecutionDisable(d) {
		d.SetId(strconv.FormatInt(rand.Int63n(math.MaxInt32), 10))
		return nil
	}

	var body = createBody(d, []string{application, workflow, pack, deploymentProfile}, []string{manualApproval, approver, schedule}, false)

	var overrides = make(map[string]interface{})
	if app, ok := d.GetOk(overridesApplication); ok {
		overrides["application"] = addSlashToNameIfMissing(app.(map[string]interface{}))
	}
	if workflow, ok := d.GetOk(overridesWorkflow); ok {
		overrides["workflow"] = addSlashToNameIfMissing(workflow.(map[string]interface{}))
	}
	if packages, ok := d.GetOk(overridesPackage); ok {
		overrides["package"] = addSlashToNameIfMissing(packages.(map[string]interface{}))
	}

	if components, ok := d.GetOk(overridesComponent); ok {
		var overridesComponentBody = make(map[string]interface{})
		for _, component := range components.([]interface{}) {
			var mapComponent = component.(map[string]interface{})
			if mapComponent["component_name"] == nil {
				return fmt.Errorf(missingRequiredFieldMessage, "component_name")
			}
			var componentName = mapComponent["component_name"].(string)
			overridesComponentBody[componentName] = addSlashToNameIfMissing(cloneMap(mapComponent, "component_name"))
		}
		overrides["components"] = overridesComponentBody
	}
	body["overrides"] = overrides

	if overrideProperty, ok := d.GetOk(overrideExistingComponents); ok {
		if overrideProperty.(bool) {
			body[propertiesMap[overrideExistingComponents]] = "OverwriteExisting"
		} else {
			body[propertiesMap[overrideExistingComponents]] = "SkipExisting"
		}
	} else {
		body[propertiesMap[overrideExistingComponents]] = "SkipExisting"
	}

	config := m.(*Config)
	response, _ := config.PostRequest("executions", body)

	executionResponse, err := convertExecutionResponse(response)
	if err != nil {
		log.Printf("[ERROR] While creating execution %s", err.Error())
		return fmt.Errorf("[ERROR] While creating execution %s", err.Error())
	}

	_ = d.Set("installation_url", executionResponse.InstallationUrl)
	_ = d.Set("monitor_url", executionResponse.MonitorUrl)
	d.SetId(strconv.FormatInt(executionResponse.Id, 10))
	return nil
}

func resourceWorkflowExecutionDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
