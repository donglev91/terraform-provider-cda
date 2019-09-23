package cda

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"net/url"
	"strings"
)

type TargetMapping struct {
	Name    string
	Targets []Entity
}

func resourceCdaDeploymentProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeploymentProfileCreate,
		Read:   resourceDeploymentProfileRead,
		Update: resourceDeploymentProfileUpdate,
		Delete: resourceDeploymentProfileDelete,

		Schema: map[string]*schema.Schema{
			name: {
				Type:     schema.TypeString,
				Required: true,
			},
			application: {
				Type:     schema.TypeString,
				Required: true,
			},
			environment: {
				Type:     schema.TypeString,
				Required: true,
			},
			loginObject: {
				Type:     schema.TypeString,
				Optional: true,
			},
			folder: {
				Type:     schema.TypeString,
				Optional: true,
			},
			description: {
				Type:     schema.TypeString,
				Optional: true,
			},
			owner: {
				Type:     schema.TypeString,
				Optional: true,
			},
			deploymentMap: {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func buildTargetMap(deploymentMap map[string]interface{}, components []BaseEntity) ([]TargetMapping, error) {
	var tm TargetMapping
	var tmList []TargetMapping

	for _, comp := range components {
		tm.Targets = make([]Entity, 0)
		if v, ok := deploymentMap[comp.Name]; ok {
			val, ok := v.(string)
			if ok && val != "" {
				tm.Name = comp.Name
				for _, target := range strings.Split(val, ",") {
					tm.Targets = append(tm.Targets, Entity{Name: strings.TrimSpace(target)})
				}

				tmList = append(tmList, tm)
			} else {
				const msg = "[ERROR] Invalid deployment target list data %s"
				log.Printf(msg, val)
				return nil, fmt.Errorf(msg, val)
			}
		} else {
			tm.Name = comp.Name
			tm.Targets = []Entity{}
			tmList = append(tmList, tm)
		}
	}

	return tmList, nil
}

func createDeploymentProfileBody(d *schema.ResourceData, config *Config) (map[string]interface{}, error) {
	var body = createBody(d, []string{name, application, environment}, []string{description, loginObject}, true)
	body, err := handlingBodyRequest(d, body, config)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func resourceDeploymentProfileCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	body, err := createDeploymentProfileBody(d, config)
	if err != nil {
		return err
	}

	err = createThenValidateResponse(d, m, DeploymentProfileType, body)
	if err != nil {
		return err
	}
	return createTargetMapping(d, config, m, true)
}

func createTargetMapping(d *schema.ResourceData, config *Config, m interface{}, firstCreate bool) error {
	targetMapping, found := d.GetOk(deploymentMap)
	if !found && firstCreate {
		return nil
	}

	components, err := getComponents(d, config)
	if err != nil {
		return err
	}

	body, err := buildTargetMap(targetMapping.(map[string]interface{}), components)
	if err != nil {
		return err
	}

	jsonStr, _ := json.Marshal(&body)
	response, _ := config.PostRequestWithByteSlice("profiles/"+d.Id()+"/target_mappings", jsonStr)

	isCreated, errorOutput, err := GetStatus(response)
	if !isCreated {
		rollBackErr := resourceDeploymentProfileDelete(d, m)
		if rollBackErr == nil {
			const msg = "[ERROR] While creating deployment profile target mapping: %s"
			if errorOutput != nil {
				log.Printf(msg, errorOutput.Error+errorOutput.Details)
				return fmt.Errorf(msg, errorOutput.Error+errorOutput.Details)
			}
			if err != nil {
				return fmt.Errorf(msg, err.Error())
			}
		} else {
			const rollbackMsg = "[ERROR] While creating deployment profile target mapping then rollback: %s"
			return fmt.Errorf(rollbackMsg, rollBackErr.Error())
		}
	}
	return nil
}

func getComponents(d *schema.ResourceData, config *Config) ([]BaseEntity, error) {
	response, err := config.GetRequest("applications?name=" + url.QueryEscape(d.Get(application).(string)))
	if err != nil || response.StatusCode != 200 {
		log.Printf("[ERROR] Get Request failed")
		return nil, err
	}

	listResponse, err := convertListResponse(response)
	if err != nil {
		log.Printf("[ERROR] Convert Get application by name response failed")
		return nil, err
	}
	appData := listResponse.Data
	if appData == nil || len(appData) == 0 {
		return nil, fmt.Errorf("[ERROR] Cannot find the application with name %s", d.Get(application).(string))
	}

	var id = appData[0].Id
	var request = fmt.Sprintf("applications/%d/components", id)
	rep, err := config.GetRequest(request)
	if err != nil || rep.StatusCode != 200 {
		log.Printf("[ERROR] Get Request failed")
		return nil, err
	}

	repList, err := convertListResponse(rep)
	if err != nil {
		log.Printf("[ERROR] Convert Get components by application id response failed")
		return nil, err
	}

	compData := repList.Data
	if compData == nil || len(compData) == 0 {
		return []BaseEntity{}, nil
	}

	return compData, nil
}

func resourceDeploymentProfileRead(d *schema.ResourceData, m interface{}) error {
	return readThenValidateResponse(d, m, DeploymentProfileType)
}

func resourceDeploymentProfileUpdate(d *schema.ResourceData, m interface{}) error {
	if d.Id() == "" {
		return fmt.Errorf("[ERROR] Deployment profile does not exist")
	}

	var body = createUpdateBody(d, []string{name, folder, description, owner, application, environment, loginObject}, true)
	err := updateThenValidateResponse(d, m, DeploymentProfileType, body)
	if err != nil {
		return err
	}
	if _, ok := isUpdateNeed(d, deploymentMap); ok {
		config := m.(*Config)
		err = createTargetMapping(d, config, m, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceDeploymentProfileDelete(d *schema.ResourceData, m interface{}) error {
	_ = resourceDeploymentProfileRead(d, m)
	return deleteThenValidateResponse(d, m, DeploymentProfileType)
}
