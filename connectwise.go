package connectwise

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// SystemInfo returned from Connectwise - GET /system/info
type SystemInfo struct {
	Version        string `json:"version"`
	IsCloud        bool   `json:"isCloud"`
	ServerTimeZone string `json:"serverTimeZone"`
	CloudRegion    string `json:"cloudRegion"`
}

// APIVersion is info from connectwise to help us create the correct base url
type APIVersion struct {
	CompanyName string `json:"CompanyName"`
	Codebase    string `json:"Codebase"`
	VersionCode string `json:"VersionCode"`
	CompanyID   string `json:"CompanyID"`
	IsCloud     bool   `json:"IsCloud"`
	SiteURL     string `json:"SiteUrl"`
}

// CwClient is a 'holder' struct for everything needed to authenticate to cw api
type CwClient struct {
	APIVersion APIVersion
	ClientID   string
	CompanyID  string
	PublicKey  string
	PrivateKey string
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
func (c CwClient) Post(path string, payload []byte, options ...CwOption) (string, error) {
	baseURL := fmt.Sprintf("https://api-na.myconnectwise.net/%sapis/3.0", c.APIVersion.Codebase)
	url := fmt.Sprintf("%s/%s", baseURL, path)
	client := &http.Client{}

	// Setup the post request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	// Header setup
	// set client id
	req.Header.Set("ClientID", c.ClientID)
	// set authorization base64(companyid+public:private)
	auth := fmt.Sprintf("%s+%s:%s", c.CompanyID, c.PublicKey, c.PrivateKey)
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

// GetAll returns all results by following pagination
func (c CwClient) GetAll(path string, options ...CwOption) (jsonPages []string, err error) {
	var currentPage int = 1
	var results []string

	for {
		page := CwOption{Key: "page", Value: strconv.Itoa(currentPage)}
		newOpts := append(options, page)
		resp, err := c.Get(path, newOpts...)
		if err != nil {
			return jsonPages, err
		}

		// Some requests will do normal pagination
		// So when they get to the end they return an empty setjjj
		if string(resp) == "[]" {
			break
		}

		// Some endpoints (like /system/info) do not follow normal
		// pagination, they just return the same data twice
		// so if we get that then just return lastResponse
		if len(results) > 0 && string(resp) == string(results[0]) {
			return []string{string(resp)}, nil
		}

		results = append(results, string(resp))
		currentPage++
	}

	return results, nil
}

// Get is an api primitive to get data from the connectwise api
func (c CwClient) Get(path string, options ...CwOption) (jsonData []byte, err error) {
	baseURL := fmt.Sprintf("https://api-na.myconnectwise.net/%sapis/3.0", c.APIVersion.Codebase)
	url := fmt.Sprintf("%s/%s", baseURL, path)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return jsonData, err
	}

	/// Header Authentication
	// set client id
	req.Header.Set("ClientID", c.ClientID)
	// set authorization base64(companyid+public:private)
	auth := fmt.Sprintf("%s+%s:%s", c.CompanyID, c.PublicKey, c.PrivateKey)
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
func NewCwClient(site string, clientID string, company string, publicKey string, privateKey string) (cwclient CwClient, err error) {
	apiVersion, err := GetAPIVersion(site, company)
	if err != nil {
		return cwclient, fmt.Errorf("Cannot get apiversion for %s at %s: %w", company, site, err)
	}
	cwclient = CwClient{
		APIVersion: apiVersion,
		ClientID:   clientID,
		CompanyID:  company,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
	return
}

// GetAPIVersion will dynamically get the Api version for this client, all that
// is required is the site and company, no authentication is needed at this point
func GetAPIVersion(site string, company string) (version APIVersion, err error) {
	url := fmt.Sprintf("https://%s/login/companyinfo/%s", site, company)
	// #nosec - gosec will detect this as a G107 error
	// the point of this is to request a "variable" url
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &version)
	if err != nil {
		return version, fmt.Errorf("Can't get unmarshal data %w", err)
	}

	return
}
