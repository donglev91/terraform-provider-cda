package cda

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
)

func resourceCdaLoginObject() *schema.Resource {
	return &schema.Resource{
		Create: resourceLoginObjectCreate,
		Read:   resourceLoginObjectRead,
		Update: resourceLoginObjectUpdate,
		Delete: resourceLoginObjectDestroy,

		Schema: map[string]*schema.Schema{
			name: {
				Type:     schema.TypeString,
				Required: true,
			},
			description: {
				Type:     schema.TypeString,
				Optional: true,
			},
			folder: {
				Type:     schema.TypeString,
				Optional: true,
			},
			owner: {
				Type:     schema.TypeString,
				Optional: true,
			},
			credentials: {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func resourceLoginObjectCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	body, err := createLoginObjectBody(d, config)

	if err != nil {
		return err
	}

	err = createThenValidateResponse(d, m, LoginObjectType, body)
	if err != nil {
		return err
	}
	return resourceLoginObjectCredentialsCreate(d, config)
}

func resourceLoginObjectRead(d *schema.ResourceData, m interface{}) error {
	return readThenValidateResponse(d, m, LoginObjectType)
}

func resourceLoginObjectUpdate(d *schema.ResourceData, m interface{}) error {
	if d.Id() == "" {
		return fmt.Errorf("[ERROR] Login object does not exist")
	}
	var body = createUpdateBody(d, []string{name, folder, description, owner}, false)
	err := updateThenValidateResponse(d, m, LoginObjectType, body)
	if err != nil {
		return err
	}

	config := m.(*Config)
	err = resourceLoginObjectCredentialsDelete(d, config)
	if err != nil {
		log.Printf("[ERROR] Delete current credentials request for re-creating new ones failed")
		return err
	}

	err = resourceLoginObjectCredentialsCreate(d, config)
	if err != nil {
		log.Printf("[ERROR] Post credentials request failed")
		return err
	}

	return nil
}

func resourceLoginObjectDestroy(d *schema.ResourceData, m interface{}) error {
	_ = resourceLoginObjectRead(d, m)
	return deleteThenValidateResponse(d, m, LoginObjectType)
}

func createLoginObjectBody(d *schema.ResourceData, config *Config) (map[string]interface{}, error) {
	var body = createBody(d, []string{name}, []string{}, false)
	body, err := handlingBodyRequest(d, body, config)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func resourceLoginObjectCredentialsCreate(d *schema.ResourceData, config *Config) error {
	for _, credential := range d.Get(credentials).([]interface{}) {
		if credential != nil {
			var credentialMap = credential.(map[string]interface{})
			body, err := createLoginObjectCredentialsBody(credentialMap)
			if err != nil {
				return err
			}
			response, _ := config.PostRequest("logins/"+d.Id()+"/credentials", body)
			_, _, errorOutput, err := Status(response)
			const msg = "[ERROR] While creating login object credentials %s"
			if errorOutput != nil {
				log.Printf(msg, errorOutput.Error+errorOutput.Details)
				return fmt.Errorf(msg, errorOutput.Error+errorOutput.Details)
			}
			if err != nil {
				return fmt.Errorf(msg, err.Error())
			}
		} else {
			return fmt.Errorf(missingRequiredFieldMessage, "credential")
		}
	}
	return nil
}

func resourceLoginObjectCredentialsDelete(d *schema.ResourceData, config *Config) error {
	var loginObjectId = d.Id()
	credentialsResponse, _ := config.GetRequest("logins/" + loginObjectId + "/credentials")
	listResponse, err := convertListCredentialsResponse(credentialsResponse)
	if err != nil {
		return err
	} else {
		for _, credential := range listResponse.Data {
			_, deleteErr := config.DeleteRequest("logins/" + loginObjectId + "/credentials/" + credential.Identify)
			if deleteErr != nil {
				return deleteErr
			}
		}
	}
	return nil
}

func createLoginObjectCredentialsBody(credential map[string]interface{}) (map[string]interface{}, error) {
	var body map[string]interface{}
	body = make(map[string]interface{})

	var errstrings []string

	if credential[agent] == nil {
		errstrings = append(errstrings, fmt.Errorf(missingRequiredFieldMessage, agent).Error())
	} else {
		body[propertiesMap[name]] = credential[agent]
	}

	if credential[customType] == nil {
		errstrings = append(errstrings, fmt.Errorf(missingRequiredFieldMessage, customType).Error())
	} else {
		body[customType] = credential[customType]
	}

	if credential[username] == nil {
		errstrings = append(errstrings, fmt.Errorf(missingRequiredFieldMessage, username).Error())
	} else {
		body[propertiesMap[username]] = credential[username]
	}

	if credential[password] == nil {
		errstrings = append(errstrings, fmt.Errorf(missingRequiredFieldMessage, password).Error())
	} else {
		body[propertiesMap[password]] = credential[password]
	}

	if errstrings == nil {
		return body, nil
	} else {
		return body, fmt.Errorf(strings.Join(errstrings, "\n"))
	}
}
