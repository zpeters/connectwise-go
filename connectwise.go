// Connectwise-go is a simple API helper for the Connectwise Manage API
package connectwise

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// System Info returned from Connectwise
//	GET /system/info
type SystemInfo struct {
	Version        string        `json:"version"`
	IsCloud        bool          `json:"isCloud"`
	ServerTimeZone string        `json:"serverTimeZone"`
	LicenseBits    []LicenseBits `json:"licenseBits"`
	CloudRegion    string        `json:"cloudRegion"`
}

// LicenseBits is a sub field of the SystemInfo request
type LicenseBits struct {
	Name       string `json:"name"`
	ActiveFlag string `json:"activeFlag"`
}

// ApiVersion is info from connectwise to help us create the correct base url
// https://developer.connectwise.com/Best_Practices/Manage_Cloud_URL_Formatting?mt-learningpath=manage
type ApiVersion struct {
	CompanyName string `json:"CompanyName"`
	Codebase    string `json:"Codebase"`
	VersionCode string `json:"VersionCode"`
	CompanyID   string `json:"CompanyID"`
	IsCloud     bool   `json:"IsCloud"`
	SiteUrl     string `json:"SiteUrl"`
}

// CwClient is a 'holder' struct for everything needed to authenticate to cw api
type CwClient struct {
	ApiVersion ApiVersion
	clientId   string
	companyId  string
	publicKey  string
	privateKey string
}

// CwOption makes up one (of multiple) options that we can pass to function
// Example: Setting the page size to 10
//		cw := CwOption{Key: "pagesize", Value: "10"}
type CwOption struct {
	Key   string
	Value string
}

// GetSystemInfo will retrieve the system info from connectwise
func (c CwClient) GetSystemInfo(options ...CwOption) (info SystemInfo, err error) {
	j, err := c.Get("/system/info", options...)
	if err != nil {
		return info, fmt.Errorf("Can't get system info %w", err)
	}
	err = json.Unmarshal(j, &info)
	if err != nil {
		return info, fmt.Errorf("Can't get unmarshal data %w", err)
	}
	return info, nil
}

// Post is an api primitive to get data from the connectwise api
func (c CwClient) Post(path string, payload interface{}, options ...CwOption) (string, error) {
	baseUrl := fmt.Sprintf("https://api-na.myconnectwise.net/%sapis/3.0", c.ApiVersion.Codebase)
	url := fmt.Sprintf("%s/%s", baseUrl, path)
	client := &http.Client{}

	// Convert payload to json
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Setup the post request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	/// Header setup
	// set client id
	req.Header.Set("ClientID", c.clientId)
	// set authorization base64(companyid+public:private)
	auth := fmt.Sprintf("%s+%s:%s", c.companyId, c.publicKey, c.privateKey)
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encoded))
	// content type
	req.Header.Set("Content-Type", "application/json")

	/// query parameters, if any
	if len(options) > 0 {
		q := req.URL.Query()
		for _, opt := range options {
			q.Add(opt.Key, opt.Value)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("Non-201 status: Code: %d Status: %s Message: %s", resp.StatusCode, resp.Status, body)
	}
	return string(body), nil
}

// Get is an api primitive to get data from the connectwise api
func (c CwClient) Get(path string, options ...CwOption) (jsonData []byte, err error) {
	baseUrl := fmt.Sprintf("https://api-na.myconnectwise.net/%sapis/3.0", c.ApiVersion.Codebase)
	url := fmt.Sprintf("%s/%s", baseUrl, path)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return jsonData, err
	}

	/// Header Authentication
	// set client id
	req.Header.Set("ClientID", c.clientId)
	// set authorization base64(companyid+public:private)
	auth := fmt.Sprintf("%s+%s:%s", c.companyId, c.publicKey, c.privateKey)
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encoded))

	/// query parameters, if any
	if len(options) > 0 {
		q := req.URL.Query()
		for _, opt := range options {
			q.Add(opt.Key, opt.Value)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return jsonData, err
	}
	defer resp.Body.Close()

	jsonData, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return jsonData, err
	}

	if resp.StatusCode != 200 {
		return jsonData, fmt.Errorf("Non-200 status: Code: %d Status: %s Message: %s", resp.StatusCode, resp.Status, jsonData)
	}
	return jsonData, nil
}

// NewCwClient creates a new client
func NewCwClient(site string, clientId string, company string, publicKey string, privateKey string) (cwclient CwClient, err error) {
	apiVersion, err := GetApiVersion(site, company)
	if err != nil {
		return
	}
	cwclient = CwClient{
		ApiVersion: apiVersion,
		clientId:   clientId,
		companyId:  company,
		publicKey:  publicKey,
		privateKey: privateKey,
	}
	return
}

// GetApiVersion will dynamically get the Api version for this client, all that
// is required is the site and company, no authentication is needed at this point
func GetApiVersion(site string, company string) (version ApiVersion, err error) {
	url := fmt.Sprintf("https://%s/login/companyinfo/%s", site, company)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	json.Unmarshal(body, &version)

	return
}
