package cda

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

// Config ... CDA configuration details
type Config struct {
	CdaServer string
	User      string
	Password  string
	Folder    string
	Owner     string
}

//checkConnection ... checkConnection with server.
func (c *Config) checkConnection() (*Config, error) {
	cdaUrl := c.CdaServer + "/api/system/v1/version_info"
	log.Printf("[DEBUG] Checking url")
	_, err := url.Parse(cdaUrl)
	if err != nil {
		log.Println("[Error] URL is not in correct format")
		return nil, err
	}
	log.Printf("[DEBUG] Creating request")
	request, err := http.NewRequest("GET", cdaUrl, nil)
	if err != nil {
		log.Printf("[ERROR] Error in creating http Request %s", err)
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	request.SetBasicAuth(c.User, c.Password)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Transport: tr}
	log.Printf("[DEBUG] Checking users")
	resp, err := client.Do(request)
	if err != nil {
		log.Println(" [ERROR] Connecting to server ", err)
		return nil, fmt.Errorf("[ERROR] CdaUrl is incorrect")

	}
	if resp.Status == "200 OK" {
		return c, nil
	}
	return nil, fmt.Errorf("[ERROR] Incorrect User or Password %s", err)
}
