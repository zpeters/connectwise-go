# Connectwise
![Go Build](https://github.com/zpeters/stashbox/workflows/Go%20Build/badge.svg)
![Go Test](https://github.com/zpeters/stashbox/workflows/Go%20Test/badge.svg)
![Go Lint](https://github.com/zpeters/stashbox/workflows/Go%20Lint/badge.svg)
![Gosec](https://github.com/zpeters/stashbox/workflows/Gosec/badge.svg)
![CodeQL](https://github.com/zpeters/stashbox/workflows/CodeQL/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/zpeters/connectwise-go)](https://goreportcard.com/report/github.com/zpeters/connectwise-go)
[![License](https://img.shields.io/github/license/zpeters/connectwise-go)](https://img.shields.io/github/license/zpeters/connectwise-go)

**Version** v0.0.3

*Connectwise* is a very simple api helper for the Connectwise Manage API.  It is meant firstly to just avoid some code duplication. It does **not** come close to covering the entire API, but it does help get you started

# Prerequisites
- Manage Public Key (See Below)
- Manage Private Key (See Below)
- Manage ClientID (See Below)

# Obtaining Keys and Client id 
## Pub and Private keys
- System -> Members
- API members tab (create a new user - if none exists - at the lowest required permissions level)
- Open the API member
- Click the API Keys tab
- Click new, give the key a name
- Record the Pub and Private key in a secure location
## Client ID
- this is required during the api authentication and calls
- get it from https://developer.connectwise.com/ClientID
- one per each app/environment (dev, prod)
- this gets added to the header as:

# Low and High level commands
In this module there is the idea of "low" and "high" level commands.  The low level commands are our basic "GET", "POST", etc.  The high level commands wrap these, along with calling the proper URL, etc. Only *a small portion* of low or high level commands are implemented at this point. Below are a few examples, see the docs for the most up-to-date details


| Level | Command | Parameters | Returns |
|-------|---------|---------------|------|
| Low   | Get     | path, options | JSON | 
| Low   | Post     | path, payload, options | JSON | 
| High  | GetSystemInfo | _NA_ | SystemInfo |

## Pagination and Retrys
Currently there is no pagination or retry mechanism (though it will be added in the future) . For the "low" level commands you can pass a "pagesize" option and manually retrieve the results page by page.  For the "high" level commands the pagesize is automatically maxed-out (1000 pages)

Occassionaly, there may be a retryable error (system timeout, etc).  At the present anything other than a `200` response for `GET`s and a `201` for a `POST` is considered an error.  Eventually, there will be a more robust mechanism for this.
  
# Examples
**TBD**
- NewCwClient
``` go
site := "na.myconnectwise.net"
clientid := "123-123-123-2134"
company := "myco"
publicKey := "my-public-key" 
privateKey := "my-private-key" 
client, err := NewCwClient(site, clientid, company, publicKey, privateKey)
...
```
- Get
``` go
client, err := NewCwClient(site, clientid, company, publicKey, privateKey)
if err != nil {
  // do something
 }
 jsonResponseString, err := client.Get("/system/info")
 ...
```
- Post
``` go
client, err := NewCwClient(site, clientid, company, publicKey, privateKey)
if err != nil {
  // do something
 }
 
 // this could be a raw string our you could
 // encode it from a struct
activityJSONPayload := []byte("{name: 'Test Post for Connectwise Go Unit Test', assignTo: { identifier: 'aMember'}}") 
resp, err := tc.client.Post("/sales/activities, activityJSONPayload)
if err != nil {
  // do something
 }
 
 // possibly do someting with 'resp'
 // it is the response from the server of a (hopefully)
 // successful post
 fmt.Println(resp)
...
```
- GetSystemInfo
``` go
  // create our client as above...
  resp, err := client.GetSystemInfo()
  if err != nil {
  // do something
 }
 
 // resp is a SystemInfo struct
 fmt.Println("Version: ", resp.Version)
```
- Passing CwOptions
```go
client, err := NewCwClient(site, clientid, company, publicKey, privateKey)
if err != nil {
  // do something
 }
 
 // get all members
 allMembers, err := client.Get("/system/members")
 if err != nil {
  // do something
 }
 
 // get a few members
 pageSize := CwOption{"Key": "pagesize", "Value": "3"}
  someMembers, err := client.Get("/system/members", pageSize)
 if err != nil {
  // do something
 }
 
 ...
```

# Contributing
Please see [CONTRIBUTING.md](CONTRIBUTING.md)

# License
Please see [LICENSE](LICENSE)
