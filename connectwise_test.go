package connectwise

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestGetApiVersion(t *testing.T) {
	tests := []struct {
		site             string
		company          string
		wantedAPIVersion APIVersion
		wantedError      error
	}{
		{"na.myconnectwise.net", "abcdef", APIVersion{
			CompanyName: "abcdef",
			Codebase:    "v2020_3/",
			VersionCode: "v2020.4",
			CompanyID:   "abcdef",
			IsCloud:     true,
			SiteURL:     "api-na.myconnectwise.net",
		}, nil},
		{"staging.connectwisedev.com", "abcdef", APIVersion{
			CompanyName: "abcdef",
			Codebase:    "v2020_3/",
			VersionCode: "v2020.4",
			CompanyID:   "abcdef",
			IsCloud:     true,
			SiteURL:     "api-staging.connectwisedev.com",
		}, nil},
		{"fake.fake.com", "abcdef", APIVersion{}, errors.New("Get \"https://fake.fake.com/login/companyinfo/abcdef\": dial tcp: lookup fake.fake.com: no such host")},
	}

	for _, tc := range tests {
		got, err := GetAPIVersion(tc.site, tc.company)
		if tc.wantedError == nil {
			require.NoError(t, err)
			require.Equal(t, tc.wantedAPIVersion, got)
		} else {
			require.EqualError(t, err, tc.wantedError.Error())
			require.Equal(t, tc.wantedAPIVersion, got)
		}
	}
}

func TestNewCwClient(t *testing.T) {
	// invalid credentials
	var invalidSite = "abcdef"
	var invalidClientID = "12345"
	var invalidCompany = "12345"
	var invalidPublicKey = "12345"
	var invalidPrivateKey = "12345"

	// load local testing credentials - read from .env
	// This could be an error if we are loading from the environment instead
	_ = godotenv.Load()
	var validSite = os.Getenv("TEST_SITE")
	var validClientID = os.Getenv("TEST_CLIENTID")
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
		{invalidSite, invalidClientID, invalidCompany, invalidPublicKey, invalidPrivateKey, CwClient{}, errors.New("Cannot get apiversion for 12345 at abcdef: Get \"https://abcdef/login/companyinfo/12345\"")},
		{validSite, validClientID, validCompany, validPublicKey, validPrivateKey, CwClient{
			APIVersion: APIVersion{
				CompanyName: validCompany,
				Codebase:    "v2020_3/",
				VersionCode: "v2020.3",
				CompanyID:   "buscominc",
				IsCloud:     true,
				SiteURL:     "api-na.myconnectwise.net",
			},

			CompanyID:  validCompany,
			ClientID:   validClientID,
			PublicKey:  validPublicKey,
			PrivateKey: validPrivateKey,
		}, nil},
	}
	for _, tc := range tests {
		got, err := NewCwClient(tc.site, tc.clientid, tc.company, tc.pubKey, tc.privKey)
		if tc.wantedError == nil {
			require.NoError(t, err)
			require.Equal(t, tc.wantedClient, got)
		} else {
			require.Contains(t, err.Error(), tc.wantedError.Error())
			require.Equal(t, tc.wantedClient, got)
		}
	}
}

func TestGetAll(t *testing.T) {
	// valid testing credentials - read from .env
	// This could be an error if we are loading from the environment instead
	_ = godotenv.Load()
	var validSite = os.Getenv("TEST_SITE")
	var validClientID = os.Getenv("TEST_CLIENTID")
	var validCompany = os.Getenv("TEST_COMPANY")
	var validPublicKey = os.Getenv("TEST_PUBKEY")
	var validPrivateKey = os.Getenv("TEST_PRIVKEY")

	// create a good client
	cwClient, err := NewCwClient(validSite, validClientID, validCompany, validPublicKey, validPrivateKey)
	require.NoError(t, err)

	var tests = []struct {
		inputPath     string
		pageSize      CwOption
		expected      string
		expectedError error
	}{
		{"/system/info", CwOption{}, "version", nil},
		{"/system/members", CwOption{}, "zpeters", nil},
		{"/system/members", CwOption{}, "bhatten", nil},
		// Make sure we force our results into pages
		{"/system/members", CwOption{Key: "pagesize", Value: "5"}, "zpeters", nil},
		{"/system/members", CwOption{Key: "pagesize", Value: "7"}, "bhatten", nil},
	}

	for _, tt := range tests {
		got, err := cwClient.GetAll(tt.inputPath, tt.pageSize)
		if tt.expectedError == nil {
			require.NoError(t, err)
		} else {
			require.EqualError(t, err, tt.expectedError.Error())
		}
		// join our slice of strings together for easier
		// searching
		allGot := strings.Join(got, "")
		require.Contains(t, allGot, tt.expected)
	}
}

func TestGetAllComparedToGet(t *testing.T) {
	// valid testing credentials - read from .env
	// This could be an error if we are loading from the environment instead
	_ = godotenv.Load()
	var validSite = os.Getenv("TEST_SITE")
	var validClientID = os.Getenv("TEST_CLIENTID")
	var validCompany = os.Getenv("TEST_COMPANY")
	var validPublicKey = os.Getenv("TEST_PUBKEY")
	var validPrivateKey = os.Getenv("TEST_PRIVKEY")

	// create a good client
	cwClient, err := NewCwClient(validSite, validClientID, validCompany, validPublicKey, validPrivateKey)
	require.NoError(t, err)

	var tests = []struct {
		inputPath string
		pageSize  CwOption
	}{
		{"/system/info", CwOption{}},
		// Max out the page size so we know that "normal" Get will get everything
		{"/system/members", CwOption{Key: "pagesize", Value: "1000"}},
		{"/sales/activities/statuses", CwOption{Key: "pagesize", Value: "1000"}},
		{"/system/kpiCategories", CwOption{Key: "pagesize", Value: "1000"}},
		{"/system/myCompany/other", CwOption{Key: "pagesize", Value: "1000"}},
	}

	for _, tt := range tests {
		gotGet, err1 := cwClient.Get(tt.inputPath, tt.pageSize)
		gotGetAll, err2 := cwClient.GetAll(tt.inputPath, tt.pageSize)

		require.NoError(t, err1)
		require.NoError(t, err2)

		// merge slice result together - normally
		// we page through the results and decode the json
		gotGetAllJoined := strings.Join(gotGetAll, "")

		require.Equal(t, string(gotGet), gotGetAllJoined)
	}
}

func TestGetSystemInfo(t *testing.T) {
	// valid testing credentials - read from .env
	// This could be an error if we are loading from the environment instead
	_ = godotenv.Load()
	var validSite = os.Getenv("TEST_SITE")
	var validClientID = os.Getenv("TEST_CLIENTID")
	var validCompany = os.Getenv("TEST_COMPANY")
	var validPublicKey = os.Getenv("TEST_PUBKEY")
	var validPrivateKey = os.Getenv("TEST_PRIVKEY")

	// create a good client
	cwClient, err := NewCwClient(validSite, validClientID, validCompany, validPublicKey, validPrivateKey)
	require.NoError(t, err)

	tests := []struct {
		client             CwClient
		wantedVersionMajor string
		wantedCloud        bool
		wantedTimeZone     string
		wantedRegion       string
		wantedError        error
	}{
		{cwClient, "v2020", true, "Eastern Standard Time", "NA", nil},
	}
	for _, tt := range tests {
		got, err := tt.client.GetSystemInfo()
		if tt.wantedError == nil {
			require.NoError(t, err)
		} else {
			require.EqualError(t, err, tt.wantedError.Error())
		}
		require.Equal(t, tt.wantedVersionMajor, strings.Split(got.Version, ".")[0])
		require.Equal(t, tt.wantedCloud, got.IsCloud)
		require.Equal(t, tt.wantedTimeZone, got.ServerTimeZone)
		require.Equal(t, tt.wantedRegion, got.CloudRegion)
	}
}

func TestGet(t *testing.T) {
	// valid testing credentials - read from .env
	// This could be an error if we are loading from the environment instead
	_ = godotenv.Load()
	var validSite = os.Getenv("TEST_SITE")
	var validClientID = os.Getenv("TEST_CLIENTID")
	var validCompany = os.Getenv("TEST_COMPANY")
	var validPublicKey = os.Getenv("TEST_PUBKEY")
	var validPrivateKey = os.Getenv("TEST_PRIVKEY")

	// create a good client
	cwClient, err := NewCwClient(validSite, validClientID, validCompany, validPublicKey, validPrivateKey)
	require.NoError(t, err)

	tests := []struct {
		client         CwClient
		path           string
		options        []CwOption
		wantedContains string
		wantedError    error
	}{
		{cwClient, "/system/info", nil, "isCloud", nil},
		{cwClient, "/system/members", []CwOption{{Key: "pagesize", Value: "1"}}, "identifier", nil},
		{cwClient, "/system/members", []CwOption{{Key: "pagesize", Value: "5"}, {Key: "page", Value: "2"}}, "identifier", nil},
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
	// This could be an error if we are loading from the environment instead
	_ = godotenv.Load()
	var validSite = os.Getenv("TEST_SITE")
	var validClientID = os.Getenv("TEST_CLIENTID")
	var validCompany = os.Getenv("TEST_COMPANY")
	var validPublicKey = os.Getenv("TEST_PUBKEY")
	var validPrivateKey = os.Getenv("TEST_PRIVKEY")

	// valid activity json snippet
	activityJSON := []byte("{name: 'Test Post for Connectwise Go Unit Test', assignTo: { identifier: 'zpeters'}}")

	// create a good client
	cwClient, err := NewCwClient(validSite, validClientID, validCompany, validPublicKey, validPrivateKey)
	require.NoError(t, err)

	tests := []struct {
		client         CwClient
		path           string
		payload        []byte
		options        []CwOption
		wantedContains string
		wantedError    error
	}{
		{cwClient, "/sales/activities", activityJSON, nil, "Test Post for Connectwise Go Unit Test", nil},
	}
	for _, tc := range tests {
		got, err := tc.client.Post(tc.path, tc.payload, tc.options...)
		if tc.wantedError == nil {
			require.NoError(t, err)
			require.Contains(t, string(got), tc.wantedContains)
		} else {
			require.EqualError(t, err, tc.wantedError.Error())
			require.Contains(t, string(got), tc.wantedContains)
		}
	}
}
