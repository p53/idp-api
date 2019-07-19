package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/p53/idp-api/apierror"
	"gotest.tools/assert"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestFailHttpClientDoRequest(t *testing.T) {
	baseClient := &http.Client{}
	apiClient := &APIClient{baseClient}
	byteArr := []byte("")
	testConfig := getUnitTestConfig()
	req, _ := http.NewRequest("POST", testConfig.IdpURL, bytes.NewBuffer(byteArr))

	_, err := apiClient.doRequest(req)

	if err == nil {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestSuccessRequestDoRequest(t *testing.T) {
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))

	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "/test")
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}

	_, err := apiClient.doRequest(req)

	if err != nil {
		t.Fatalf("Method should return success, error is: %s!", err)
	}
}

func TestFailedRequestDoRequest(t *testing.T) {
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))

	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "/test")
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}

	_, err := apiClient.doRequest(req)

	if err == nil {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestInvalidBasicAuthHeadersGetAuthBodyFromBasicAuth(t *testing.T) {
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()
	_, _, err := getAuthBodyFromBasicAuth(rr, req, controller)

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestAuthStringsGetAuthBodyFromBasicAuth(t *testing.T) {
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	req.SetBasicAuth("test", "test")

	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()
	authArr, url, err := getAuthBodyFromBasicAuth(rr, req, controller)

	if err != nil {
		t.Fatalf("Function should not fail! %s", err)
	}

	if len(authArr) != 2 {
		t.Fatalf("There should be two auth strings! currently there are %d", len(authArr))
	}

	if ok := strings.Contains(url, testConfig.IdpRealm); !ok {
		t.Fatalf("There should be master domain in url: %s", url)
	}
}

func TestAuthStringsGetAdminAuthBody(t *testing.T) {
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	req.SetBasicAuth("test", "test")

	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()
	authArr, url, err := getAdminAuthBody(rr, req, controller)

	if err != nil {
		t.Fatalf("Function should not fail! %s", err)
	}

	if len(authArr) != 1 {
		t.Fatalf("There should be one auth string! currently there are %d", len(authArr))
	}

	if ok := strings.Contains(url, "master"); !ok {
		t.Fatalf("There should be master domain in url: %s", url)
	}
}

func TestInvalidBasicAuthHeadersAuthenticate(t *testing.T) {
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	authAdminFunc := getAuthBodyFromBasicAuth
	apiClient := &APIClient{BaseClient: &http.Client{}}
	_, _, err := apiClient.authenticate(rr, req, controller, authAdminFunc)

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestInvalidRequestAuthenticate(t *testing.T) {
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	req.SetBasicAuth("test", "test")
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	authAdminFunc := getAuthBodyFromBasicAuth
	apiClient := &APIClient{BaseClient: &http.Client{}}
	_, _, err := apiClient.authenticate(rr, req, controller, authAdminFunc)

	if err == nil {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}

	if rr.Result().StatusCode != 401 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestSuccessfulAuthenticate(t *testing.T) {
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	req.SetBasicAuth("test", "test")
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	testClientURL := fmt.Sprintf(testConfig.TokenURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(testTokenBody)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	authAdminFunc := getAuthBodyFromBasicAuth
	apiClient := &APIClient{BaseClient: testClient}
	token, _, err := apiClient.authenticate(rr, req, controller, authAdminFunc)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	if rr.Result().StatusCode != 200 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}

	if token != "test_access_token" {
		t.Fatalf("Bad token %s", token)
	}
}

func TestFailureExtractingTokenAuthenticate(t *testing.T) {
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	req.SetBasicAuth("test", "test")
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	testClientURL := fmt.Sprintf(testConfig.TokenURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(testBadTokenBody)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	authAdminFunc := getAuthBodyFromBasicAuth
	apiClient := &APIClient{BaseClient: testClient}
	_, _, err := apiClient.authenticate(rr, req, controller, authAdminFunc)

	if err == nil {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}

	if rr.Result().StatusCode != 500 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestFailureCreateClient(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	testClientURL := fmt.Sprintf(testConfig.ClientsURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	err := apiClient.createClient(rr, controller, "test_token", Client{})

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}

	if rr.Result().StatusCode != 500 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestSuccessCreateClient(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	testClientURL := fmt.Sprintf(testConfig.ClientsURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 201,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	err := apiClient.createClient(rr, controller, "test_token", Client{})

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	if rr.Result().StatusCode != 200 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestFailureGetClientId(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	testClientURL := fmt.Sprintf(testConfig.ClientsURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	_, err := apiClient.getClientID(rr, controller, "test_token", ClientWithSecret{})

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}

	if rr.Result().StatusCode != 500 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestJqFailureGetClientId(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	testClientURL := fmt.Sprintf(testConfig.ClientsURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	_, err := apiClient.getClientID(rr, controller, "test_token", ClientWithSecret{})

	if err == nil {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}

	if rr.Result().StatusCode != 500 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestJqSuccessGetClientId(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	testClientURL := fmt.Sprintf(testConfig.ClientsURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(testClientsData)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	inputClient := ClientWithSecret{
		ClientID: "test",
	}

	_, err := apiClient.getClientID(rr, controller, "test_token", inputClient)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	if rr.Result().StatusCode != 200 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestFailureGetClient(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	testClientURL := fmt.Sprintf(testConfig.ClientsURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	_, err := apiClient.getClient(rr, controller, "test_token", Client{})

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestJqFailureGetClient(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	testClientURL := fmt.Sprintf(testConfig.ClientsURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	_, err := apiClient.getClient(rr, controller, "test_token", Client{})

	if err == nil {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestJqSuccessGetClient(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	testClientURL := fmt.Sprintf(testConfig.ClientsURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(testClientsData)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}

	_, err := apiClient.getClient(rr, controller, "test_token", Client{})

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}
}

func TestFailureGetClientSecret(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	clientUID := "40b5444c-5990-496d-bb67-64c535df8dc4"
	testClientURL := fmt.Sprintf(testConfig.ClientSecretURI, testConfig.IdpURL, testConfig.IdpRealm, clientUID)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	_, err := apiClient.getClientSecret(rr, controller, "test_token", clientUID)

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}

	if rr.Result().StatusCode != 500 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestSuccessGetClientSecret(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	clientUID := "40b5444c-5990-496d-bb67-64c535df8dc4"
	testClientURL := fmt.Sprintf(testConfig.ClientSecretURI, testConfig.IdpURL, testConfig.IdpRealm, clientUID)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(testClientSecret)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	secret, err := apiClient.getClientSecret(rr, controller, "test_token", clientUID)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	if rr.Result().StatusCode != 200 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}

	if secret != "test_secret" {
		t.Fatalf("Bad secret value %s", secret)
	}
}

func TestFailureUpdateClient(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	clientUID := "40b5444c-5990-496d-bb67-64c535df8dc4"
	testClientURL := fmt.Sprintf(testConfig.ClientURI, testConfig.IdpURL, testConfig.IdpRealm, clientUID)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	err := apiClient.updateClient(rr, controller, "test_token", Client{}, clientUID)

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}

	if rr.Result().StatusCode != 500 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestSuccessUpdateClient(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	clientUID := "40b5444c-5990-496d-bb67-64c535df8dc4"
	testClientURL := fmt.Sprintf(testConfig.ClientURI, testConfig.IdpURL, testConfig.IdpRealm, clientUID)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(testClientSecret)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	err := apiClient.updateClient(rr, controller, "test_token", Client{}, clientUID)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	if rr.Result().StatusCode != 200 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestFailureDeleteClient(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	clientUID := "40b5444c-5990-496d-bb67-64c535df8dc4"
	testClientURL := fmt.Sprintf(testConfig.ClientURI, testConfig.IdpURL, testConfig.IdpRealm, clientUID)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	err := apiClient.deleteClient(rr, controller, "test_token", clientUID)

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}

	if rr.Result().StatusCode != 500 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestSuccessDeleteClient(t *testing.T) {
	testConfig := getUnitTestConfig()
	controller := &Controller{Config: testConfig}
	rr := httptest.NewRecorder()

	clientUID := "40b5444c-5990-496d-bb67-64c535df8dc4"
	testClientURL := fmt.Sprintf(testConfig.ClientURI, testConfig.IdpURL, testConfig.IdpRealm, clientUID)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(testClientSecret)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	err := apiClient.deleteClient(rr, controller, "test_token", clientUID)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	if rr.Result().StatusCode != 200 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestFailureCreateUser(t *testing.T) {
	testConfig := getUnitTestConfig()

	testClientURL := fmt.Sprintf(testConfig.UsersURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	err := apiClient.createUser(testConfig, "test_token", &User{})

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestSuccessCreateUser(t *testing.T) {
	testConfig := getUnitTestConfig()

	testClientURL := fmt.Sprintf(testConfig.UsersURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 201,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	err := apiClient.createUser(testConfig, "test_token", &User{})

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}
}

func TestFailureGetUserId(t *testing.T) {
	testConfig := getUnitTestConfig()

	testClientURL := fmt.Sprintf(testConfig.UsersURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	_, err := apiClient.getUserID(testConfig, "test_token", &User{})

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestJqFailureGetUserId(t *testing.T) {
	testConfig := getUnitTestConfig()

	testClientURL := fmt.Sprintf(testConfig.UsersURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	_, err := apiClient.getUserID(testConfig, "test_token", &User{})

	if err == nil {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestJqSuccessGetUserId(t *testing.T) {
	testConfig := getUnitTestConfig()

	testClientURL := fmt.Sprintf(testConfig.UsersURI, testConfig.IdpURL, testConfig.IdpRealm)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(testUsersData)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	inputUser := User{
		Username: "test",
	}

	_, err := apiClient.getUserID(testConfig, "test_token", &inputUser)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}
}

func TestFailureDeleteUser(t *testing.T) {
	testConfig := getUnitTestConfig()

	userUID := "40b5444c-5990-496d-bb67-64c535df8dc4"
	testClientURL := fmt.Sprintf(testConfig.UserURI, testConfig.IdpURL, testConfig.IdpRealm, userUID)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	err := apiClient.deleteUser(testConfig, "test_token", userUID)

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestSuccessDeleteUser(t *testing.T) {
	testConfig := getUnitTestConfig()
	rr := httptest.NewRecorder()

	userUID := "40b5444c-5990-496d-bb67-64c535df8dc4"
	testClientURL := fmt.Sprintf(testConfig.UserURI, testConfig.IdpURL, testConfig.IdpRealm, userUID)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString("")),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	err := apiClient.deleteUser(testConfig, "test_token", userUID)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	if rr.Result().StatusCode != 200 {
		t.Fatalf("Bad return code %d", rr.Result().StatusCode)
	}
}

func TestFailureSetUserPassword(t *testing.T) {
	testConfig := getUnitTestConfig()

	userUID := "40b5444c-5990-496d-bb67-64c535df8dc4"
	testClientURL := fmt.Sprintf(testConfig.UserPasswordURI, testConfig.IdpURL, testConfig.IdpRealm, userUID)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`FAIL`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	credential := &UserSecret{Type: "password", Value: "test"}
	err := apiClient.setUserPassword(testConfig, "test_token", credential, userUID)

	if _, ok := err.(*apierror.ApiError); !ok {
		t.Fatalf("Method doesn't fail when it should! %s", err)
	}
}

func TestSuccessSetUserPassword(t *testing.T) {
	testConfig := getUnitTestConfig()
	userUID := "40b5444c-5990-496d-bb67-64c535df8dc4"
	testClientURL := fmt.Sprintf(testConfig.UserPasswordURI, testConfig.IdpURL, testConfig.IdpRealm, userUID)
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testClientURL)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString("")),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	apiClient := &APIClient{BaseClient: testClient}
	credential := &UserSecret{Type: "password", Value: "test"}
	err := apiClient.setUserPassword(testConfig, "test_token", credential, userUID)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}
}
