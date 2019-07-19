package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/p53/idp-api/apierror"
)

func getUnitTestConfig() *Config {
	config := &Config{
		IdpURL:          "https://fake.com",
		ClientID:        "fake",
		ClientSecret:    "fake",
		ApiClientID:     "test",
		ApiClientSecret: "test",
		IdpAdmin:        "fake",
		IdpPass:         "fake",
		IdpRealm:        "master",
		CheckURI:        "%s/auth/admin",
		ClientsURI:      "%s/auth/admin/realms/%s/clients",
		ClientURI:       "%s/auth/admin/realms/%s/clients/%s",
		ClientSecretURI: "%s/auth/admin/realms/%s/clients/%s/client-secret",
		TokenURI:        "%s/auth/realms/%s/protocol/openid-connect/token",
		UsersURI:        "%s/auth/admin/realms/%s/users",
		UserURI:         "%s/auth/admin/realms/%s/users/%s",
		UserPasswordURI: "%s/auth/admin/realms/%s/users/%s/reset-password",
	}

	return config
}

func getFuncTestConfig() *Config {
	config := &Config{
		IdpURL:          "http://keycloak-server:8080",
		ClientID:        "test",
		ClientSecret:    "fake",
		ApiClientID:     "admin-cli",
		ApiClientSecret: "test",
		IdpAdmin:        "admin",
		IdpPass:         "admin",
		IdpRealm:        "master",
		CheckURI:        "%s/auth/admin",
		ClientsURI:      "%s/auth/admin/realms/%s/clients",
		ClientURI:       "%s/auth/admin/realms/%s/clients/%s",
		ClientSecretURI: "%s/auth/admin/realms/%s/clients/%s/client-secret",
		TokenURI:        "%s/auth/realms/%s/protocol/openid-connect/token",
		UsersURI:        "%s/auth/admin/realms/%s/users",
		UserURI:         "%s/auth/admin/realms/%s/users/%s",
		UserPasswordURI: "%s/auth/admin/realms/%s/users/%s/reset-password",
	}

	return config
}

func TestSwagger(t *testing.T) {
	apiClient := &APIClientMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte("")

	req, err := http.NewRequest("GET", "/swagger.yml", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/swagger.yml", ctrl.ReadSwagger).Methods("GET")
	r.ServeHTTP(rr, req)

	if rr.Code != 200 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}
}

func TestHealth(t *testing.T) {
	apiClient := &APIClientMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte("")

	req, err := http.NewRequest("GET", "/health", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/health", ctrl.HealthCheck).Methods("GET")
	r.ServeHTTP(rr, req)

	if rr.Code != 200 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}
}

func TestIdpErrorHealth(t *testing.T) {
	apiClient := &APIClientInternalServerErrorMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte("")

	req, err := http.NewRequest("GET", "/health", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/health", ctrl.HealthCheck).Methods("GET")
	r.ServeHTTP(rr, req)

	if rr.Code != 500 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}
}

func TestCreateClient(t *testing.T) {
	apiClient := &APIClientMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte(testPayload)

	req, err := http.NewRequest("POST", "/client", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/client", ctrl.CreateResource).Methods("POST")
	r.ServeHTTP(rr, req)

	if rr.Code != 201 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}

	secret := &ClientSecret{}

	secErr := json.Unmarshal(rr.Body.Bytes(), secret)

	if secErr != nil {
		t.Fatalf("Problem unmarshalling %s", secErr)
	}

	if secret.Value != "testsecret" {
		t.Fatalf("Value of secret is wrong %s", secret.Value)
	}
}

func TestMissingRequiredFieldsPayloadCreateUser(t *testing.T) {
	apiClient := &APIClientMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte(testBadPayload)

	req, err := http.NewRequest("POST", "/client", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/client", ctrl.CreateResource).Methods("POST")
	r.ServeHTTP(rr, req)

	if rr.Code != 400 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}

	retErr := &apierror.ApiError{}
	errAPI := json.Unmarshal([]byte(rr.Body.String()), retErr)

	if errAPI != nil {
		t.Fatal("Problem unmarshalling error")
	}

	if retErr.Code != "1007" {
		t.Fatal(fmt.Sprintf("Wrong apierror code %s", retErr.Code))
	}
}

func TestInvalidRequestPayloadCreateUser(t *testing.T) {
	apiClient := &APIClientMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte("[bad_payload]")

	req, err := http.NewRequest("POST", "/client", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/client", ctrl.CreateResource).Methods("POST")
	r.ServeHTTP(rr, req)

	if rr.Code != 400 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}

	retErr := &apierror.ApiError{}
	errAPI := json.Unmarshal([]byte(rr.Body.String()), retErr)

	if errAPI != nil {
		t.Fatal("Problem unmarshalling error")
	}

	if retErr.Code != "1003" {
		t.Fatal(fmt.Sprintf("Wrong apierror code %s", retErr.Code))
	}
}

func TestIdpErrorCreateUser(t *testing.T) {
	apiClient := &APIClientInternalServerErrorMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte(testPayload)

	req, err := http.NewRequest("POST", "/client", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/client", ctrl.CreateResource).Methods("POST")
	r.ServeHTTP(rr, req)

	if rr.Code != 500 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}
}

func TestUpdateClient(t *testing.T) {
	apiClient := &APIClientMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte(testSecretPayload)

	req, err := http.NewRequest("PUT", "/client", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/client", ctrl.UpdateResource).Methods("PUT")
	r.ServeHTTP(rr, req)

	if rr.Code != 201 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}
}

func TestMissingRequiredFieldsPayloadUpdateUser(t *testing.T) {
	apiClient := &APIClientMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte(testPayload)

	req, err := http.NewRequest("PUT", "/client", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/client", ctrl.UpdateResource).Methods("PUT")
	r.ServeHTTP(rr, req)

	if rr.Code != 400 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}

	retErr := &apierror.ApiError{}
	errAPI := json.Unmarshal([]byte(rr.Body.String()), retErr)

	if errAPI != nil {
		t.Fatal("Problem unmarshalling error")
	}

	if retErr.Code != "1007" {
		t.Fatal(fmt.Sprintf("Wrong apierror code %s", retErr.Code))
	}
}

func TestInvalidRequestPayloadUpdateUser(t *testing.T) {
	apiClient := &APIClientMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte("[bad_payload]")

	req, err := http.NewRequest("PUT", "/client", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/client", ctrl.UpdateResource).Methods("PUT")
	r.ServeHTTP(rr, req)

	if rr.Code != 400 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}

	retErr := &apierror.ApiError{}
	errAPI := json.Unmarshal([]byte(rr.Body.String()), retErr)

	if errAPI != nil {
		t.Fatal("Problem unmarshalling error")
	}

	if retErr.Code != "1003" {
		t.Fatal(fmt.Sprintf("Wrong apierror code %s", retErr.Code))
	}
}

func TestIdpErrorUpdateUser(t *testing.T) {
	apiClient := &APIClientInternalServerErrorMock{}
	testConfig := getUnitTestConfig()
	ctrl := &Controller{Config: testConfig}
	testConfig.HTTPClient = apiClient

	payload := []byte(testSecretPayload)

	req, err := http.NewRequest("PUT", "/client", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	r.HandleFunc("/client", ctrl.UpdateResource).Methods("PUT")
	r.ServeHTTP(rr, req)

	if rr.Code != 500 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}
}
