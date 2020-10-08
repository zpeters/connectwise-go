package connectwise

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestGetApiVersion(t *testing.T) {
	tests := []struct {
		site             string
		company          string
		wantedApiVersion ApiVersion
		wantedError      error
	}{
		{"na.myconnectwise.net", "abcdef", ApiVersion{
			CompanyName: "abcdef",
			Codebase:    "v2020_3/",
			VersionCode: "v2020.3",
			CompanyID:   "abcdef",
			IsCloud:     true,
			SiteUrl:     "api-na.myconnectwise.net",
		}, nil},
		{"staging.connectwisedev.com", "abcdef", ApiVersion{
			CompanyName: "abcdef",
			Codebase:    "v2020_3/",
			VersionCode: "v2020.4",
			CompanyID:   "abcdef",
			IsCloud:     true,
			SiteUrl:     "api-staging.connectwisedev.com",
		}, nil},
		{"fake.fake.com", "abcdef", ApiVersion{}, errors.New("Get \"https://fake.fake.com/login/companyinfo/abcdef\": dial tcp: lookup fake.fake.com: no such host")},
	}

	for _, tc := range tests {
		got, err := GetApiVersion(tc.site, tc.company)
		if tc.wantedError == nil {
			require.NoError(t, err)
			require.Equal(t, tc.wantedApiVersion, got)
		} else {
			require.EqualError(t, err, tc.wantedError.Error())
			require.Equal(t, tc.wantedApiVersion, got)
		}
	}
}

func TestNewCwClient(t *testing.T) {
	// invalid credentials
	var invalidSite = "abcdef"
	var invalidClientId = "12345"
	var invalidCompany = "12345"
	var invalidPublicKey = "12345"
	var invalidPrivateKey = "12345"

	// valid testing credentials - read from .env
	err := godotenv.Load()
	require.NoError(t, err)
	var validSite = os.Getenv("TEST_SITE")
	var validClientId = os.Getenv("TEST_CLIENTID")
	var validCompany = os.Getenv("TEST_COMPANY")
	var validPublicKey = os.Getenv("TEST_PUBKEY")
	var validPrivateKey = os.Getenv("TEST_PRIVKEY")

	tests := []struct {
		site         string
		clientid     string
		company      string
		pubKey       string
		privKey      string
		wantedClient CwClient
		wantedError  error
	}{
		{invalidSite, invalidClientId, invalidCompany, invalidPublicKey, invalidPrivateKey, CwClient{}, errors.New("Cannot get apiversion for 12345 at abcdef: Get \"https://abcdef/login/companyinfo/12345\": dial tcp: lookup abcdef: no such host")},
		{validSite, validClientId, validCompany, validPublicKey, validPrivateKey, CwClient{
			ApiVersion: ApiVersion{
				CompanyName: validCompany,
				Codebase:    "v2020_3/",
				VersionCode: "v2020.3",
				CompanyID:   validCompany,
				IsCloud:     true,
				SiteUrl:     "api-na.myconnectwise.net",
			},

			companyId:  validCompany,
			clientId:   validClientId,
			publicKey:  validPublicKey,
			privateKey: validPrivateKey,
		}, nil},
	}
	for _, tc := range tests {
		got, err := NewCwClient(tc.site, tc.clientid, tc.company, tc.pubKey, tc.privKey)
		if tc.wantedError == nil {
			require.NoError(t, err)
			require.Equal(t, tc.wantedClient, got)
		} else {
			require.EqualError(t, err, tc.wantedError.Error())
			require.Equal(t, tc.wantedClient, got)
		}
	}
}

func TestGetSystemInfo(t *testing.T) {
	// valid testing credentials - read from .env
	err := godotenv.Load()
	require.NoError(t, err)
	var validSite = os.Getenv("TEST_SITE")
	var validClientId = os.Getenv("TEST_CLIENTID")
	var validCompany = os.Getenv("TEST_COMPANY")
	var validPublicKey = os.Getenv("TEST_PUBKEY")
	var validPrivateKey = os.Getenv("TEST_PRIVKEY")

	// create a good client
	cwClient, err := NewCwClient(validSite, validClientId, validCompany, validPublicKey, validPrivateKey)
	require.NoError(t, err)

	tests := []struct {
		client      CwClient
		wantedInfo  SystemInfo
		wantedError error
	}{
		{cwClient, SystemInfo{
			Version:        "v2020.3.75324",
			IsCloud:        true,
			ServerTimeZone: "Eastern Standard Time",
			CloudRegion:    "NA",
		}, nil},
	}
	for _, tc := range tests {
		got, err := tc.client.GetSystemInfo()
		if tc.wantedError == nil {
			require.NoError(t, err)
			require.Equal(t, tc.wantedInfo, got)
		} else {
			require.EqualError(t, err, tc.wantedError.Error())
			require.Equal(t, tc.wantedInfo, got)
		}
	}
}

func TestGet(t *testing.T) {
	// valid testing credentials - read from .env
	err := godotenv.Load()
	require.NoError(t, err)
	var validSite = os.Getenv("TEST_SITE")
	var validClientId = os.Getenv("TEST_CLIENTID")
	var validCompany = os.Getenv("TEST_COMPANY")
	var validPublicKey = os.Getenv("TEST_PUBKEY")
	var validPrivateKey = os.Getenv("TEST_PRIVKEY")

	// create a good client
	cwClient, err := NewCwClient(validSite, validClientId, validCompany, validPublicKey, validPrivateKey)
	require.NoError(t, err)

	tests := []struct {
		client         CwClient
		path           string
		options        []CwOption
		wantedContains string
		wantedError    error
	}{
		{cwClient, "/system/info", nil, "isCloud", nil},
		{cwClient, "/system/members", []CwOption{CwOption{Key: "pagesize", Value: "1"}}, "identifier", nil},
		{cwClient, "/system/members", []CwOption{CwOption{Key: "pagesize", Value: "5"}, CwOption{Key: "page", Value: "2"}}, "identifier", nil},
	}
	for _, tc := range tests {
		got, err := tc.client.Get(tc.path, tc.options...)
		if tc.wantedError == nil {
			require.NoError(t, err)
			require.Contains(t, string(got), tc.wantedContains)
		} else {
			require.EqualError(t, err, tc.wantedError.Error())
			require.Contains(t, string(got), tc.wantedContains)
		}
	}
}

func TestPost(t *testing.T) {
	// valid testing credentials - read from .env
	err := godotenv.Load()
	require.NoError(t, err)
	var validSite = os.Getenv("TEST_SITE")
	var validClientId = os.Getenv("TEST_CLIENTID")
	var validCompany = os.Getenv("TEST_COMPANY")
	var validPublicKey = os.Getenv("TEST_PUBKEY")
	var validPrivateKey = os.Getenv("TEST_PRIVKEY")

	// valid activity json snippet
	activityJson := []byte("{\"name\":\"mything\",\"assignTo\":{\"identifier\":\"zpeters\"}}")

	// create a good client
	cwClient, err := NewCwClient(validSite, validClientId, validCompany, validPublicKey, validPrivateKey)
	require.NoError(t, err)

	tests := []struct {
		client         CwClient
		path           string
		payload        []byte
		options        []CwOption
		wantedContains string
		wantedError    error
	}{
		{cwClient, "/sales/activities", activityJson, nil, "isCloud", nil},
	}
	for _, tc := range tests {
		got, err := tc.client.Post(tc.path, tc.payload, tc.options...)
		fmt.Println(got)
		if tc.wantedError == nil {
			require.NoError(t, err)
			require.Contains(t, string(got), tc.wantedContains)
		} else {
			require.EqualError(t, err, tc.wantedError.Error())
			require.Contains(t, string(got), tc.wantedContains)
		}
	}
}
