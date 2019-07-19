package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/p53/idp-api/logging"
)

var app *App

func TestMain(m *testing.M) {
	testConfig := getFuncTestConfig()
	os.Setenv("IDP_URL", testConfig.IdpURL)
	os.Setenv("CLIENT_ID", testConfig.ClientID)
	os.Setenv("CLIENT_SECRET", testConfig.ClientSecret)
	os.Setenv("API_CLIENT_ID", testConfig.ApiClientID)
	os.Setenv("API_CLIENT_SECRET", testConfig.ApiClientSecret)
	os.Setenv("IDP_ADMIN_USER", testConfig.IdpAdmin)
	os.Setenv("IDP_ADMIN_PASSWORD", testConfig.IdpPass)
	os.Setenv("IDP_REALM", testConfig.IdpRealm)
	app = CreateApp()
	code := m.Run()
	os.Exit(code)
}

func getClientSecret(clientJson string) (clientSecret string) {
	logger := logging.GetLogger()

	apiClient := &APIClient{&http.Client{}}
	testConfig := getFuncTestConfig()
	testConfig.HTTPClient = apiClient
	controller := &Controller{Config: testConfig}
	byteArr := []byte("")
	reqf, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	rrSec := httptest.NewRecorder()

	authBodyFunc := getAdminAuthBody
	token, _, err := apiClient.authenticate(rrSec, reqf, controller, authBodyFunc)

	if err != nil {
		logger.Fatalf("Problem authenticating %s", err)
	}

	testNewClientStruct := &Client{}
	errNc := json.Unmarshal([]byte(clientJson), testNewClientStruct)

	if errNc != nil {
		logger.Fatalf("Problem unmarshalling %s", errNc)
	}

	clientInfo, errGet := apiClient.getClient(rrSec, controller, token, *testNewClientStruct)

	if errGet != nil {
		logger.Fatalf("Method fail when it shouldn't! %s", errGet)
	}

	clientSecret, errSec := apiClient.getClientSecret(rrSec, controller, token, clientInfo.ID)

	if errSec != nil {
		logger.Fatalf("Method fail when it shouldn't! %s", errSec)
	}

	return clientSecret
}

func setupEndpointTest(t *testing.T) func(t *testing.T) {
	logger := logging.GetLogger()
	logger.Println("########### Setup test ############")

	apiClient := &APIClient{&http.Client{}}
	testConfig := getFuncTestConfig()
	testConfig.HTTPClient = apiClient
	controller := &Controller{Config: testConfig}
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	rr := httptest.NewRecorder()

	authBodyFunc := getAdminAuthBody
	token, _, err := apiClient.authenticate(rr, req, controller, authBodyFunc)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	newClient := &Client{}
	errUn := json.Unmarshal([]byte(testPayload), newClient)

	if errUn != nil {
		t.Fatalf("Problem unmarshalling %s", errUn)
	}

	logger.Printf("Creating test client: %s", newClient.ClientID)
	errCreate := apiClient.createClient(rr, controller, token, *newClient)

	if errCreate != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errCreate)
	}

	newUser := &User{}
	errUnU := json.Unmarshal([]byte(testUserPayload), newUser)

	if errUnU != nil {
		t.Fatalf("Problem unmarshalling %s", errUnU)
	}

	logger.Printf("Creating test user: %s", newUser.Username)
	errCreateU := apiClient.createUser(testConfig, token, newUser)

	if errCreateU != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errCreateU)
	}

	logger.Printf("Getting test user id: %s", newUser.Username)
	userID, errGetU := apiClient.getUserID(testConfig, token, newUser)

	if errGetU != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errGetU)
	}

	newCredential := &UserSecret{}
	errUnC := json.Unmarshal([]byte(testUserSecretPayload), newCredential)

	if errUnC != nil {
		t.Fatalf("Problem unmarshalling %s", errUnC)
	}

	logger.Printf("Setting up test user %s credentianls", newUser.Username)
	errReset := apiClient.setUserPassword(testConfig, token, newCredential, userID)

	if errReset != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errReset)
	}

	logger.Println("########### End of setup ############")

	return func(t *testing.T) {
		logger.Println("########### Teardown test ###########")

		newClientWithSecret := &Client{}
		errUnm := json.Unmarshal([]byte(testPayload), newClientWithSecret)

		if errUnm != nil {
			t.Fatalf("Problem unmarshalling %s", errUnm)
		}

		logger.Printf("Getting client id for client %s", newClient.ClientID)
		clientInfo, errGet := apiClient.getClient(rr, controller, token, *newClientWithSecret)

		if errGet != nil {
			t.Fatalf("Method fail when it shouldn't! %s", errGet)
		}

		logger.Printf("Delete client: %s", newClient.ClientID)
		errDelete := apiClient.deleteClient(rr, controller, token, clientInfo.ID)

		if errDelete != nil {
			t.Fatalf("Method fail when it shouldn't! %s", errDelete)
		}

		logger.Printf("Delete user: %s with id %s", newClient.ClientID, userID)
		errDeleteU := apiClient.deleteUser(testConfig, token, userID)

		if errDeleteU != nil {
			t.Fatalf("Method fail when it shouldn't! %s", errDeleteU)
		}

		logger.Println("########### End of Teardown test ###########")
	}
}

func setupClientAuthEndpointTest(t *testing.T) func(t *testing.T) {
	logger := logging.GetLogger()
	logger.Println("########### Setup test ############")

	apiClient := &APIClient{&http.Client{}}
	testConfig := getFuncTestConfig()
	testConfig.HTTPClient = apiClient
	controller := &Controller{Config: testConfig}
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	rr := httptest.NewRecorder()

	authBodyFunc := getAdminAuthBody
	token, _, err := apiClient.authenticate(rr, req, controller, authBodyFunc)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	newClient := &Client{}
	errUn := json.Unmarshal([]byte(testPayload), newClient)

	if errUn != nil {
		t.Fatalf("Problem unmarshalling %s", errUn)
	}

	newServiceClient := &Client{}
	errUns := json.Unmarshal([]byte(testServiceClient), newServiceClient)

	if errUns != nil {
		t.Fatalf("Problem unmarshalling %s", errUns)
	}

	clientsSlice := []*Client{newClient, newServiceClient}

	for _, item := range clientsSlice {
		logger.Printf("Creating test client: %s", item.ClientID)
		errCreate := apiClient.createClient(rr, controller, token, *item)

		if errCreate != nil {
			t.Fatalf("Method fail when it shouldn't! %s", errCreate)
		}
	}

	logger.Println("########### End of setup ############")

	return func(t *testing.T) {
		logger.Println("########### Teardown test ###########")

		for _, item := range clientsSlice {
			logger.Printf("Getting client id for client %s", item.ClientID)
			clientInfo, errGet := apiClient.getClient(rr, controller, token, *item)

			if errGet != nil {
				t.Fatalf("Method fail when it shouldn't! %s", errGet)
			}

			logger.Printf("Delete client: %s", item.ClientID)
			errDelete := apiClient.deleteClient(rr, controller, token, clientInfo.ID)

			if errDelete != nil {
				t.Fatalf("Method fail when it shouldn't! %s", errDelete)
			}
		}

		logger.Println("########### End of Teardown test ###########")
	}
}

func TestIntegrationSwagger(t *testing.T) {
	apiClient := &APIClient{&http.Client{}}
	testConfig := getFuncTestConfig()
	testConfig.HTTPClient = apiClient
	byteArr := []byte("")

	req, err := http.NewRequest("GET", "/swagger.yml", bytes.NewBuffer(byteArr))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.router.ServeHTTP(rr, req)

	if rr.Code != 200 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}
}

func TestIntegrationHealth(t *testing.T) {
	apiClient := &APIClient{&http.Client{}}
	testConfig := getFuncTestConfig()
	testConfig.HTTPClient = apiClient
	byteArr := []byte("")

	req, err := http.NewRequest("GET", "/health", bytes.NewBuffer(byteArr))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.router.ServeHTTP(rr, req)

	if rr.Code != 200 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}
}

func TestIntegrationAdminAuthenticate(t *testing.T) {
	apiClient := &APIClient{&http.Client{}}
	testConfig := getFuncTestConfig()
	testConfig.HTTPClient = apiClient
	controller := &Controller{Config: testConfig}
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	rr := httptest.NewRecorder()

	authBodyFunc := getAdminAuthBody
	token, _, err := apiClient.authenticate(rr, req, controller, authBodyFunc)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	var tokenIntf interface{} = token

	if val, ok := tokenIntf.(string); !ok {
		t.Fatalf("Token is not string! %s", val)
	}
}

func TestIntegrationCreateDeleteClient(t *testing.T) {
	apiClient := &APIClient{&http.Client{}}
	testConfig := getFuncTestConfig()
	testConfig.HTTPClient = apiClient
	controller := &Controller{Config: testConfig}
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	rr := httptest.NewRecorder()

	authBodyFunc := getAdminAuthBody
	token, _, err := apiClient.authenticate(rr, req, controller, authBodyFunc)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	newClient := &Client{}
	errUn := json.Unmarshal([]byte(testPayload), newClient)

	if errUn != nil {
		t.Fatalf("Problem unmarshalling %s", errUn)
	}

	newClientWithSecret := &Client{}
	errUnm := json.Unmarshal([]byte(testPayload), newClientWithSecret)

	if errUnm != nil {
		t.Fatalf("Problem unmarshalling %s", errUnm)
	}

	errCreate := apiClient.createClient(rr, controller, token, *newClient)

	if errCreate != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errCreate)
	}

	clientOut, errGet := apiClient.getClient(rr, controller, token, *newClientWithSecret)

	if errGet != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errGet)
	}

	errDelete := apiClient.deleteClient(rr, controller, token, clientOut.ID)

	if errDelete != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errDelete)
	}
}

func TestIntegrationCreateDeleteUser(t *testing.T) {
	apiClient := &APIClient{&http.Client{}}
	testConfig := getFuncTestConfig()
	testConfig.HTTPClient = apiClient
	controller := &Controller{Config: testConfig}
	byteArr := []byte("")
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(byteArr))
	rr := httptest.NewRecorder()

	authBodyFunc := getAdminAuthBody
	token, _, err := apiClient.authenticate(rr, req, controller, authBodyFunc)

	if err != nil {
		t.Fatalf("Method fail when it shouldn't! %s", err)
	}

	newUser := &User{}
	errUn := json.Unmarshal([]byte(testUserPayload), newUser)

	if errUn != nil {
		t.Fatalf("Problem unmarshalling %s", errUn)
	}

	errCreate := apiClient.createUser(testConfig, token, newUser)

	if errCreate != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errCreate)
	}

	userID, errGet := apiClient.getUserID(testConfig, token, newUser)

	if errGet != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errGet)
	}

	newCredential := &UserSecret{}
	errUnC := json.Unmarshal([]byte(testUserSecretPayload), newCredential)

	if errUnC != nil {
		t.Fatalf("Problem unmarshalling %s", errUnC)
	}

	errReset := apiClient.setUserPassword(testConfig, token, newCredential, userID)

	if errReset != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errReset)
	}

	errDelete := apiClient.deleteUser(testConfig, token, userID)

	if errDelete != nil {
		t.Fatalf("Method fail when it shouldn't! %s", errDelete)
	}
}

// TestIntegrationUserAuthEndpoint - tests create/update/delete client with
// user account authentication
func TestIntegrationUserAuthEndpoint(t *testing.T) {
	tearDownEndpointTest := setupEndpointTest(t)
	defer tearDownEndpointTest(t)
	logger := logging.GetLogger()

	testUser := &User{}
	testUserPass := &UserSecret{}

	errUnU := json.Unmarshal([]byte(testUserPayload), testUser)

	if errUnU != nil {
		t.Fatalf("Problem unmarshalling %s", errUnU)
	}

	errUnUs := json.Unmarshal([]byte(testUserSecretPayload), testUserPass)

	if errUnUs != nil {
		t.Fatalf("Problem unmarshalling %s", errUnUs)
	}

	logger.Println("Test auth failure creating client")
	reqFail, errFail := http.NewRequest("POST", "/api/v1/client", bytes.NewBufferString(testNewClient))
	reqFail.SetBasicAuth("bad", "bad")

	if errFail != nil {
		t.Fatal(errFail)
	}

	rrFail := httptest.NewRecorder()
	app.router.ServeHTTP(rrFail, reqFail)

	if rrFail.Code != 401 {
		content := rrFail.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rrFail.Code, content))
	}

	logger.Println("Creating client")
	req, err := http.NewRequest("POST", "/api/v1/client", bytes.NewBufferString(testNewClient))
	req.SetBasicAuth(testUser.Username, testUserPass.Value)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.router.ServeHTTP(rr, req)

	if rr.Code != 201 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}

	logger.Println("Updating client")
	clientSecret := getClientSecret(testNewClient)

	clientWithSecret := &ClientWithSecret{}
	errCl := json.Unmarshal([]byte(testUpdateClient), clientWithSecret)

	if errCl != nil {
		t.Fatalf("Problem unmarshalling %s", errCl)
	}

	clientWithSecret.Secret = clientSecret

	testUpdateClientSec, merr := json.Marshal(clientWithSecret)

	if merr != nil {
		t.Fatalf("Problem marshalling %s", merr)
	}

	reqUp, err := http.NewRequest("PUT", "/api/v1/client", bytes.NewBuffer(testUpdateClientSec))
	reqUp.SetBasicAuth(testUser.Username, testUserPass.Value)

	rrUp := httptest.NewRecorder()
	app.router.ServeHTTP(rrUp, reqUp)

	if rrUp.Code != 201 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rrUp.Code, content))
	}

	logger.Println("Deleting client")
	reqDel, err := http.NewRequest("DELETE", "/api/v1/client", bytes.NewBuffer(testUpdateClientSec))
	reqDel.SetBasicAuth(testUser.Username, testUserPass.Value)

	rrDel := httptest.NewRecorder()
	app.router.ServeHTTP(rrDel, reqDel)

	if rrDel.Code != 201 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rrDel.Code, content))
	}
}

func TestIntegrationClientAuthEndpoint(t *testing.T) {
	tearDownEndpointTest := setupClientAuthEndpointTest(t)
	defer tearDownEndpointTest(t)
	logger := logging.GetLogger()

	serviceClient := &Client{}

	errUnU := json.Unmarshal([]byte(testServiceClient), serviceClient)

	if errUnU != nil {
		t.Fatalf("Problem unmarshalling %s", errUnU)
	}

	secret := getClientSecret(testServiceClient)

	logger.Println("Creating client")
	req, err := http.NewRequest("POST", "/api/v1/client", bytes.NewBufferString(testNewClient))
	req.SetBasicAuth(serviceClient.ClientID, secret)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.router.ServeHTTP(rr, req)

	if rr.Code != 201 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rr.Code, content))
	}

	logger.Println("Updating client")

	clientSecret := getClientSecret(testUpdateClient)
	clientWithSecret := &ClientWithSecret{}
	errCl := json.Unmarshal([]byte(testUpdateClient), clientWithSecret)

	if errCl != nil {
		t.Fatalf("Problem unmarshalling %s", errCl)
	}

	clientWithSecret.Secret = clientSecret

	testUpdateClientSec, merr := json.Marshal(clientWithSecret)

	if merr != nil {
		t.Fatalf("Problem marshalling %s", merr)
	}

	reqUp, err := http.NewRequest("PUT", "/api/v1/client", bytes.NewBuffer(testUpdateClientSec))
	reqUp.SetBasicAuth(serviceClient.ClientID, secret)

	rrUp := httptest.NewRecorder()
	app.router.ServeHTTP(rrUp, reqUp)

	if rrUp.Code != 201 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rrUp.Code, content))
	}

	logger.Println("Deleting client")
	reqDel, err := http.NewRequest("DELETE", "/api/v1/client", bytes.NewBuffer(testUpdateClientSec))
	reqDel.SetBasicAuth(serviceClient.ClientID, secret)

	rrDel := httptest.NewRecorder()
	app.router.ServeHTTP(rrDel, reqDel)

	if rrDel.Code != 201 {
		content := rr.Body.String()
		t.Fatal(fmt.Sprintf("Wrong response code %d %s", rrDel.Code, content))
	}
}
