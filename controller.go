package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/p53/idp-api/apierror"
	"github.com/p53/idp-api/logging"
	validator "gopkg.in/validator.v2"
)

// Controller - controller structure
type Controller struct {
	Config *Config
	db     *sql.DB
}

// ClientOut - structure for output idp client definition - there is bug/feature? in keycloak
// when you set json with ID to keycloak it will set it as uid of object
type ClientOut struct {
	ID                        string `json:"id"`
	ClientID                  string `json:"clientId" validate:"nonzero"`
	PublicClient              bool   `json:"publicClient"`
	DirectAccessGrantsEnabled bool   `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool   `json:"serviceAccountsEnabled"`
	StandardFlowEnabled       bool   `json:"standardFlowEnabled"`
	ImplicitFlowEnabled       bool   `json:"implicitFlowEnabled"`
}

// Client - structure for input idp client definition
type Client struct {
	ClientID                  string   `json:"clientId" validate:"nonzero"`
	PublicClient              bool     `json:"publicClient"`
	RedirectUris              []string `json:"redirectUris"`
	RootUrl                   string   `json:"rootUrl"`
	WebOrigins                []string `json:"webOrigins"`
	AdminUrl                  string   `json:"adminUrl"`
	DirectAccessGrantsEnabled bool     `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool     `json:"serviceAccountsEnabled"`
	StandardFlowEnabled       bool     `json:"standardFlowEnabled"`
	ImplicitFlowEnabled       bool     `json:"implicitFlowEnabled"`
	Description               string   `json:"description"`
}

// ClientWithSecret - structure for input idp client definition, containing secret
type ClientWithSecret struct {
	ClientID                  string `json:"clientId" validate:"nonzero"`
	PublicClient              bool   `json:"publicClient"`
	DirectAccessGrantsEnabled bool   `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool   `json:"serviceAccountsEnabled"`
	StandardFlowEnabled       bool   `json:"standardFlowEnabled"`
	ImplicitFlowEnabled       bool   `json:"implicitFlowEnabled"`
	Secret                    string `json:"clientSecret" validate:"nonzero"`
}

func (controller *Controller) HealthCheck(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	httpClient := controller.Config.HTTPClient

	url := fmt.Sprintf(controller.Config.CheckURI, controller.Config.IdpURL)
	byteArr := []byte("")
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	_, errReq := httpClient.doRequest(req)

	if errReq != nil {
		logger.Println(errReq)
		errStr := fmt.Sprintf("%s", errReq)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		http.Error(w, inverr.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (controller *Controller) ReadSwagger(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	filename := "swagger.yml"

	pwd, _ := os.Getwd()
	path := pwd

	if pwd != "/" {
		path += "/"
	}

	txt, errRead := ioutil.ReadFile(path + filename)

	if errRead != nil {
		logger.Printf("Error while reading swagger file %s", errRead)
		inErr := apierror.InternalServerError()
		http.Error(w, inErr.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(txt)
}

// CreateResource method for creating client
func (controller *Controller) CreateResource(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	httpClient := controller.Config.HTTPClient
	bodyFunc := getAuthBodyFromBasicAuth

	logger.Println("Authenticating external user")

	_, authEntity, err := httpClient.authenticate(w, r, controller, bodyFunc)

	if err != nil {
		return
	}

	var client Client
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if errDec := decoder.Decode(&client); errDec != nil {
		logger.Println(errDec)
		inverr := apierror.InvalidRequestPayload()
		http.Error(w, inverr.Error(), 400)
		return
	}

	if err := validator.Validate(client); err != nil {
		logger.Println(err)
		inverr := apierror.MissingRequiredFieldsPayload()
		http.Error(w, inverr.Error(), 400)
		return
	}

	if ok := client.StandardFlowEnabled; ok {
		if len(client.RedirectUris) > 0 {
			client.RootUrl = client.RedirectUris[0]
			client.AdminUrl = client.RedirectUris[0]
			client.WebOrigins = client.RedirectUris
		}
	}

	adminBodyFunc := getAdminAuthBody

	logger.Println("Authenticating app admin user")

	token, _, err := httpClient.authenticate(w, r, controller, adminBodyFunc)

	if err != nil {
		return
	}

	client.PublicClient = false
	client.Description = fmt.Sprintf("Client created by %s", authEntity)
	err = httpClient.createClient(w, controller, token, client)

	if err != nil {
		return
	}

	clientInf, errClient := httpClient.getClient(w, controller, token, client)

	if errClient != nil {
		return
	}

	clientSec, errSec := httpClient.getClientSecret(w, controller, token, clientInf.ID)

	if errSec != nil {
		return
	}

	secOut, marSecErr := json.Marshal(ClientSecret{Value: clientSec})

	if marSecErr != nil {
		msg := fmt.Sprintf("Unmarshalling failed %s", marSecErr)
		log.Println(msg)
		inErr := apierror.InternalServerError()
		http.Error(w, inErr.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(secOut)
}

// UpdateResource method for updating client
func (controller *Controller) UpdateResource(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	httpClient := controller.Config.HTTPClient
	bodyFunc := getAuthBodyFromBasicAuth
	_, _, err := httpClient.authenticate(w, r, controller, bodyFunc)

	if err != nil {
		return
	}

	var clientWithSecret ClientWithSecret
	var client Client
	bodyBytes, errBio := ioutil.ReadAll(r.Body)

	if errBio != nil {
		logger.Println(errBio)
		http.Error(w, errBio.Error(), 500)
		return
	}

	bodyReaderClient := bytes.NewBuffer(bodyBytes)
	bodyReaderClientSec := bytes.NewBuffer(bodyBytes)

	defer r.Body.Close()
	decoder := json.NewDecoder(bodyReaderClient)

	if errDec := decoder.Decode(&client); errDec != nil {
		logger.Println(errDec)
		inverr := apierror.InvalidRequestPayload()
		http.Error(w, inverr.Error(), 400)
		return
	}

	decoderSec := json.NewDecoder(bodyReaderClientSec)

	if errDecc := decoderSec.Decode(&clientWithSecret); errDecc != nil {
		logger.Println(errDecc)
		inverr := apierror.InvalidRequestPayload()
		http.Error(w, inverr.Error(), 400)
		return
	}

	if err := validator.Validate(clientWithSecret); err != nil {
		logger.Println(err)
		inverr := apierror.MissingRequiredFieldsPayload()
		http.Error(w, inverr.Error(), 400)
		return
	}

	if ok := client.StandardFlowEnabled; ok {
		if len(client.RedirectUris) > 0 {
			client.RootUrl = client.RedirectUris[0]
			client.AdminUrl = client.RedirectUris[0]
			client.WebOrigins = client.RedirectUris
		}
	}

	adminBodyFunc := getAdminAuthBody
	token, _, err := httpClient.authenticate(w, r, controller, adminBodyFunc)

	if err != nil {
		return
	}

	client.PublicClient = false
	clientInfo, err := httpClient.getClient(w, controller, token, client)

	if err != nil {
		return
	}

	clientSecret, err := httpClient.getClientSecret(w, controller, token, clientInfo.ID)

	if err != nil {
		return
	}

	if clientWithSecret.Secret != clientSecret {
		logger.Println(err)
		inverr := apierror.BadClientSecret()
		http.Error(w, inverr.Error(), 401)
		return
	}

	err = httpClient.updateClient(w, controller, token, client, clientInfo.ID)

	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// DeleteResource method for deleting client
func (controller *Controller) DeleteResource(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	httpClient := controller.Config.HTTPClient
	bodyFunc := getAuthBodyFromBasicAuth
	_, _, err := httpClient.authenticate(w, r, controller, bodyFunc)

	if err != nil {
		return
	}

	var clientWithSecret ClientWithSecret
	var client Client
	bodyBytes, errBio := ioutil.ReadAll(r.Body)

	if errBio != nil {
		logger.Println(errBio)
		http.Error(w, errBio.Error(), 500)
		return
	}

	bodyReaderClient := bytes.NewBuffer(bodyBytes)
	bodyReaderClientSec := bytes.NewBuffer(bodyBytes)

	defer r.Body.Close()
	decoder := json.NewDecoder(bodyReaderClient)

	if errDec := decoder.Decode(&client); errDec != nil {
		logger.Println(errDec)
		inverr := apierror.InvalidRequestPayload()
		http.Error(w, inverr.Error(), 400)
		return
	}

	decoderSec := json.NewDecoder(bodyReaderClientSec)

	if errDecc := decoderSec.Decode(&clientWithSecret); errDecc != nil {
		logger.Println(errDecc)
		inverr := apierror.InvalidRequestPayload()
		http.Error(w, inverr.Error(), 400)
		return
	}

	if err := validator.Validate(client); err != nil {
		logger.Println(err)
		inverr := apierror.MissingRequiredFieldsPayload()
		http.Error(w, inverr.Error(), 400)
		return
	}

	adminBodyFunc := getAdminAuthBody
	token, _, err := httpClient.authenticate(w, r, controller, adminBodyFunc)

	if err != nil {
		return
	}

	clientInfo, err := httpClient.getClient(w, controller, token, client)

	if err != nil {
		return
	}

	clientSecret, err := httpClient.getClientSecret(w, controller, token, clientInfo.ID)

	if err != nil {
		return
	}

	if clientWithSecret.Secret != clientSecret {
		logger.Println(err)
		inverr := apierror.BadClientSecret()
		http.Error(w, inverr.Error(), 401)
		return
	}

	err = httpClient.deleteClient(w, controller, token, clientInfo.ID)

	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
