package cda

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type OutputCase struct {
	Id int64 `json:"id"`
}

type ExecutionResponse struct {
	OutputCase
	MonitorUrl      string `json:"monitor_url"`
	InstallationUrl string `json:"installation_url"`
}

type ErrorCase struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Details string `json:"details"`
}

type ListResponseOutput struct {
	HasMore bool         `json:"has_more"`
	Total   int          `json:"total"`
	Data    []BaseEntity `json:"data"`
}

type ListCredentialsResponse struct {
	Data []CredentialsResponse `json:"data"`
}

type CredentialsResponse struct {
	Identify string `json:"identify"`
}

func unmarshalResponse(response *http.Response, v interface{}) error {
	respString, err := getResponseAsString(response)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(respString), v)
}

func convertExecutionResponse(response *http.Response) (*ExecutionResponse, error) {
	data := &ExecutionResponse{}
	return data, ConvertResponse(response, data)
}

func convertListResponse(response *http.Response) (*ListResponseOutput, error) {
	data := &ListResponseOutput{}
	return data, ConvertResponse(response, data)
}

func convertListCredentialsResponse(response *http.Response) (*ListCredentialsResponse, error) {
	data := &ListCredentialsResponse{}
	return data, ConvertResponse(response, data)
}

func Status(response *http.Response) (bool, *OutputCase, *ErrorCase, error) {
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		data := &OutputCase{}
		err := unmarshalResponse(response, data)
		if err != nil {
			return false, nil, nil, err
		}
		return true, data, nil, nil
	}

	data := &ErrorCase{}
	err := unmarshalResponse(response, data)
	if err != nil {
		return false, nil, nil, err
	}
	return false, nil, data, nil
}

func ConvertResponse(response *http.Response, m interface{}) error {
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		return unmarshalResponse(response, m)
	} else {
		str, _ := getResponseAsString(response)
		return fmt.Errorf(str)
	}
}

func GetStatus(response *http.Response) (bool, *ErrorCase, error) {
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		return true, nil, nil
	}

	data := &ErrorCase{}
	err := unmarshalResponse(response, data)
	if err != nil {
		return false, nil, err
	}
	return false, data, nil
}

func getResponseAsString(response *http.Response) (string, error) {
	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("[ERROR] Reading body %s", err.Error())
		return "", fmt.Errorf("[ERROR] Reading body %s", err.Error())
	}
	return string(buf), nil
}

func createThenValidateResponse(d *schema.ResourceData, m interface{}, eType EntityType, body map[string]interface{}) error {
	config := m.(*Config)
	response, _ := config.PostRequest(routerMaps[eType], body)
	output := &OutputCase{}
	err := ConvertResponse(response, output)
	if err != nil {
		const msg = "[ERROR] While creating %s %s"
		log.Printf(msg, entityTypeNameMaps[eType], err.Error())
		return fmt.Errorf(msg, entityTypeNameMaps[eType], err.Error())
	}
	d.SetId(strconv.FormatInt(output.Id, 10))
	return nil
}

func updateThenValidateResponse(d *schema.ResourceData, m interface{}, eType EntityType, body map[string]interface{}) error {
	config := m.(*Config)
	response, err := config.PostRequest(routerMaps[eType]+"/"+d.Id(), body)
	if err != nil {
		log.Printf("[ERROR] Post request failed")
		return err
	}

	if response.StatusCode != 200 {
		response, _ := getResponseAsString(response)
		return fmt.Errorf("[ERROR] Update %s fail. %s", entityTypeNameMaps[eType], response)
	}

	return nil
}

func deleteThenValidateResponse(d *schema.ResourceData, m interface{}, eType EntityType) error {
	if d.Id() == "" {
		return fmt.Errorf("[ERROR] %s does not exist", entityTypeNameMaps[eType])
	}
	config := m.(*Config)
	response, err := config.DeleteRequest(routerMaps[eType] + "/" + d.Id())

	if err != nil {
		log.Printf("[ERROR] DELETE Request failed")
		return err
	}

	if response.StatusCode == 200 {
		d.SetId("")
	} else {
		str, err := getResponseAsString(response)
		const msg = "[ERROR] While destroying %s %s"
		if err != nil {
			return fmt.Errorf(msg, entityTypeNameMaps[eType], err.Error())
		}
		return fmt.Errorf(msg, entityTypeNameMaps[eType], d.Id()+"\n"+str)
	}
	return nil
}

func readThenValidateResponse(d *schema.ResourceData, m interface{}, eType EntityType) error {
	config := m.(*Config)
	response, err := config.GetRequest(routerMaps[eType] + "?name=" + url.QueryEscape(d.Get(name).(string)))
	if err != nil {
		log.Printf("[ERROR] Get Request failed." + err.Error())
		return err
	}
	if response.StatusCode != 200 {
		d.SetId("")
	}
	return nil
}
