package cda

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"net/http"
)

const ApiPrefix = "/api/data/v1/"

// PostRequest ...
func (c *Config) PostRequest(requestPath string, body map[string]interface{}) (*http.Response, error) {
	var bodyBuffer *bytes.Buffer
	if body != nil {
		buff, _ := json.Marshal(&body)
		var jsonStr = []byte(buff)
		log.Println("[DEBUG]------------------------------------------------- body Request " + string(jsonStr))
		bodyBuffer = bytes.NewBuffer(jsonStr)
	}
	var url = c.CdaServer + ApiPrefix + requestPath
	log.Println("[DEBUG]-------------------------------------------------- Posting data on URL " + url)
	request, err := http.NewRequest("POST", url, bodyBuffer)
	if err != nil {
		log.Printf("[ERROR] Error in creating http Request %s", err)
		return nil, fmt.Errorf("[ERROR] Error in creating request %s", err.Error())
	}
	return c.getResponse(request)
}

// PostRequest ...
func (c *Config) PostRequestWithByteSlice(requestPath string, jsonStr []byte) (*http.Response, error) {
	var bodyBuffer *bytes.Buffer
	if jsonStr != nil {
		log.Println("[DEBUG]------------------------------------------------- body Request " + string(jsonStr))
		bodyBuffer = bytes.NewBuffer(jsonStr)
	}
	var url = c.CdaServer + ApiPrefix + requestPath
	log.Println("[DEBUG]-------------------------------------------------- Posting data on URL " + url)
	request, err := http.NewRequest("POST", url, bodyBuffer)
	if err != nil {
		log.Printf("[ERROR] Error in creating http Request %s", err)
		return nil, fmt.Errorf("[ERROR] Error in creating request %s", err.Error())
	}
	return c.getResponse(request)
}

func (c *Config) DeleteRequest(requestPath string) (*http.Response, error) {
	var url = c.CdaServer + ApiPrefix + requestPath
	request, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		log.Printf("[ERROR] Error in creating http Request %s", err)
		return nil, fmt.Errorf("[ERROR] Error in creating request %s", err.Error())
	}
	return c.getResponse(request)
}

// GetRequest ... get Request to cda
func (c *Config) GetRequest(requestPath string) (*http.Response, error) {
	var url = c.CdaServer + ApiPrefix + requestPath
	log.Println("[DEBUG] Getting data on URL " + url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[ERROR] Error in creating http Request %s", err)
		return nil, fmt.Errorf("[ERROR] Error in creating request %s", err.Error())
	}
	return c.getResponse(request)
}

func (c *Config) getResponse(request *http.Request) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	request.SetBasicAuth(c.User, c.Password)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Transport: tr}
	return client.Do(request)
}

func addToBody(body map[string]interface{}, key string, property interface{}) {
	switch property.(type) {
	case string:
		body[propertiesMap[key]] = property.(string)
	case bool:
		body[propertiesMap[key]] = property.(bool)
	case map[string]interface{}:
		body[propertiesMap[key]] = adjustDeserializeJsonString(property.(map[string]interface{}))
	}
}

func createBody(d *schema.ResourceData, requiredProperties []string, optionalProperties []string, hasDynamicProperty bool) map[string]interface{} {
	var body map[string]interface{}
	body = make(map[string]interface{})

	for _, property := range requiredProperties {
		addToBody(body, property, d.Get(property))
	}

	for _, property := range optionalProperties {
		if propertyValue, found := d.GetOkExists(property); found {
			addToBody(body, property, propertyValue)
		}
	}

	if hasDynamicProperty {
		dynamic, found := d.GetOk(dynamicProperties)
		if found {
			body[propertiesMap[dynamicProperties]] = addSlashToNameIfMissing(dynamic.(map[string]interface{}))
		}
	}

	return body
}

func createUpdateBody(d *schema.ResourceData, propertyNames []string, hasDynamicProperty bool) map[string]interface{} {
	var body map[string]interface{}
	body = make(map[string]interface{})

	for _, key := range propertyNames {
		if newValue, ok := isUpdateNeed(d, key); ok {
			addToBody(body, key, newValue)
		}
	}

	if hasDynamicProperty {
		if dynamic, ok := isUpdateNeed(d, dynamicProperties); ok {
			body[propertiesMap[dynamicProperties]] = addSlashToNameIfMissing(dynamic.(map[string]interface{}))
		}
	}

	return body
}
