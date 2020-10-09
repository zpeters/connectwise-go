# Connectwise go module

**Version** v0.0.1

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
- Get
- Post
- GetSystemInfo
- Passing CwOptions

# Contributing
**TBD**
