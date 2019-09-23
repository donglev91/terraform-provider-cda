package cda

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"strings"
)

func validateMapPropertyValue(v interface{}, k string) (ws []string, errors []error) {
	if v == nil {
		return
	}
	maps := v.(map[string]interface{})

	for _, val := range maps {
		if !strings.HasPrefix(val.(string), "{") {
			continue
		}
		if json.Valid([]byte(val.(string))) {
			continue
		}
		errors = append(errors, fmt.Errorf("%s is not valid json object string", val.(string)))
	}
	return
}

func adjustDeserializeJsonString(m map[string]interface{}) map[string]interface{} {
	list := make(map[string]interface{}, len(m))
	for i, v := range m {
		var result interface{} = nil
		if strings.HasPrefix(v.(string), "{") {
			rmap := make(map[string]interface{})
			err := json.Unmarshal([]byte(v.(string)), &rmap)
			if err == nil {
				result = rmap
			}
		} else {
			result = v.(string)
		}
		list[i] = result
	}
	return list
}

func addSlashToNameIfMissing(m map[string]interface{}) map[string]interface{} {
	adjusted := adjustDeserializeJsonString(m)
	list := make(map[string]interface{}, len(m))
	for i, v := range adjusted {
		if strings.HasPrefix(i, "/") {
			list[i] = v
		} else {
			list["/"+i] = v
		}
	}
	return list
}

func isUpdateNeed(d *schema.ResourceData, attName string) (interface{}, bool) {
	if d.HasChange(attName) {
		if v, ok := d.GetOk(attName); ok {
			return v, ok
		}
	}
	return nil, false
}

func handlingBodyRequest(d *schema.ResourceData, body map[string]interface{}, config *Config) (map[string]interface{}, error) {
	if own, found := d.GetOk(owner); found {
		body[propertiesMap[owner]] = own.(string)
	} else if len(config.Owner) > 0 {
		body[propertiesMap[owner]] = config.Owner
	} else {
		body[propertiesMap[owner]] = config.User
	}

	if folderStr, found := d.GetOk(folder); found {
		body[propertiesMap[folder]] = folderStr.(string)
	} else if len(config.Folder) > 0 {
		body[propertiesMap[folder]] = config.Folder
	} else {
		return nil, fmt.Errorf(missingRequiredFieldMessage, folder)
	}
	return body, nil
}

type Entity struct {
	Name string
}

type BaseEntity struct {
	Id   int
	Name string
}
