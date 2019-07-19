package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/thedevsaddam/gojsonq"
	"github.com/p53/idp-api/apierror"
	"github.com/p53/idp-api/logging"
)

// APIClientIntf - interface for idp api client
type APIClientIntf interface {
	doRequest(req *http.Request) ([]byte, error)
	authenticate(w http.ResponseWriter, r *http.Request, controller *Controller, f AuthBodyGetter) (tokenVal string, authEntity string, err error)
	createClient(w http.ResponseWriter, controller *Controller, token string, client Client) (err error)
	getClientID(w http.ResponseWriter, controller *Controller, token string, client ClientWithSecret) (clientID string, err error)
	getClient(w http.ResponseWriter, controller *Controller, token string, client Client) (clientOut *ClientOut, err error)
	getClientSecret(w http.ResponseWriter, controller *Controller, token string, clientUID string) (clientSecret string, err error)
	updateClient(w http.ResponseWriter, controller *Controller, token string, client Client, clientUID string) (err error)
	deleteClient(w http.ResponseWriter, controller *Controller, token string, clientUID string) (err error)
}

// APIClient - type for defining idp api client
type APIClient struct {
	BaseClient *http.Client
}

// Token - type for defining token outpu
type Token struct {
	Value string `json:"access_token" validate:"nonzero"`
}

// ClientID - type for defining client id output
type ClientID struct {
	ID string `json:"id"`
}

// UserID - type for defining user id output
type UserID struct {
	ID string `json:"id"`
}

// ClientSecret - type for defining client secret output
type ClientSecret struct {
	Value string `json:"value" validate:"nonzero"`
}

// User - type for defining user in input
type User struct {
	Username string `json:"username" validate:"nonzero"`
	Enabled  bool   `json:"enabled" validate:"nonzero"`
}

// UserSecret - type for defining client secret output
type UserSecret struct {
	Type      string `json:"type" validate:"nonzero"`
	Value     string `json:"value" validate:"nonzero"`
	Temporary bool   `json:"temporary"`
}

// AuthBodyGetter - type for standardizing auth body getters
type AuthBodyGetter func(w http.ResponseWriter, r *http.Request, controller *Controller) (authBody []url.Values, authUrl string, err error)

// APIClientMock - api client mock for testing normal non-error operations
type APIClientMock struct{}

func (s *APIClientMock) doRequest(req *http.Request) ([]byte, error) {
	var empty []byte
	return empty, nil
}

func (s *APIClientMock) authenticate(
	w http.ResponseWriter,
	r *http.Request,
	controller *Controller,
	f AuthBodyGetter) (tokenVal string, authEntity string, err error) {
	return tokenVal, authEntity, nil
}

func (s *APIClientMock) createClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client Client) (err error) {
	return nil
}

func (s *APIClientMock) getClientID(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client ClientWithSecret) (clientID string, err error) {
	return "", nil
}

func (s *APIClientMock) getClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client Client) (clientOut *ClientOut, err error) {
	return &ClientOut{ID: "test"}, nil
}

func (s *APIClientMock) getClientSecret(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	clientUID string) (clientSecret string, err error) {
	return "testsecret", nil
}

func (s *APIClientMock) updateClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client Client,
	clientUID string) (err error) {
	return nil
}

func (s *APIClientMock) deleteClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	clientUID string) (err error) {
	return
}

// APIClientInternalServerErrorMock - api client mock to simulate error situations
type APIClientInternalServerErrorMock struct{}

func (s *APIClientInternalServerErrorMock) doRequest(req *http.Request) ([]byte, error) {
	var empty []byte
	return empty, errors.New("Test Idp API Failure")
}

func (s *APIClientInternalServerErrorMock) authenticate(
	w http.ResponseWriter,
	r *http.Request,
	controller *Controller,
	f AuthBodyGetter) (tokenVal string, authEntity string, err error) {
	return tokenVal, authEntity, nil
}

func (s *APIClientInternalServerErrorMock) createClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client Client) (err error) {
	http.Error(w, "Test Idp API Failure", 500)
	return errors.New("Test Idp API Failure")
}

func (s *APIClientInternalServerErrorMock) getClientID(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client ClientWithSecret) (clientID string, err error) {
	return "", nil
}

func (s *APIClientInternalServerErrorMock) getClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client Client) (clientOut *ClientOut, err error) {
	return &ClientOut{ID: "test"}, nil
}

func (s *APIClientInternalServerErrorMock) getClientSecret(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	clientUID string) (clientSecret string, err error) {
	return "testsecret", nil
}

func (s *APIClientInternalServerErrorMock) updateClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client Client,
	clientUID string) (err error) {
	http.Error(w, "Test Idp API Failure", 500)
	return errors.New("Test Idp API Failure")
}

func (s *APIClientInternalServerErrorMock) deleteClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	clientUID string) (err error) {
	return
}

func (s *APIClient) doRequest(req *http.Request) ([]byte, error) {
	logger := logging.GetLogger()
	resp, err := s.BaseClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if 200 != resp.StatusCode && 201 != resp.StatusCode && 204 != resp.StatusCode {
		logger.Printf("Response code from URL: %s is %d", req.URL, resp.StatusCode)
		logger.Printf(string(body))
		msg := fmt.Sprintf("%s", body)
		return nil, errors.New(msg)
	}

	return body, nil
}

func getAdminAuthBody(
	w http.ResponseWriter,
	r *http.Request,
	controller *Controller) (authBody []url.Values, authUrl string, err error) {

	authClientCredentialAdminBody := url.Values{
		"username":      {controller.Config.IdpAdmin},
		"password":      {controller.Config.IdpPass},
		"grant_type":    {"password"},
		"client_id":     {controller.Config.ApiClientID},
		"client_secret": {controller.Config.ApiClientSecret},
	}

	authBody = []url.Values{authClientCredentialAdminBody}
	authUrl = fmt.Sprintf(controller.Config.TokenURI, controller.Config.IdpURL, "master")

	return authBody, authUrl, nil
}

func getAuthBodyFromBasicAuth(
	w http.ResponseWriter,
	r *http.Request,
	controller *Controller) (authBody []url.Values, authUrl string, err error) {
	logger := logging.GetLogger()
	username, password, ok := r.BasicAuth()

	if !ok {
		authHedErr := apierror.InvalidBasicAuthHeaders()
		logger.Println(authHedErr.Error())
		http.Error(w, authHedErr.Error(), 401)
		return nil, "", authHedErr
	}

	authResourceOwnerBody := url.Values{
		"username":      {username},
		"password":      {password},
		"grant_type":    {"password"},
		"client_id":     {controller.Config.ClientID},
		"client_secret": {controller.Config.ClientSecret},
	}

	authClientBody := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {username},
		"client_secret": {password},
	}

	authBody = []url.Values{authResourceOwnerBody, authClientBody}
	authUrl = fmt.Sprintf(controller.Config.TokenURI, controller.Config.IdpURL, controller.Config.IdpRealm)

	return authBody, authUrl, nil
}

func (s *APIClient) authenticate(
	w http.ResponseWriter,
	r *http.Request,
	controller *Controller,
	f AuthBodyGetter) (tokenVal string, authEntity string, err error) {
	logger := logging.GetLogger()

	authBody, url, err := f(w, r, controller)

	if err != nil {
		return "", "", err
	}

	var authErr error
	var tokenBody []byte

	for _, authBodyItem := range authBody {
		form := strings.NewReader(authBodyItem.Encode())
		logger.Println(url)
		req, err := http.NewRequest("POST", url, form)

		if err != nil {
			logger.Println(err)
			http.Error(w, err.Error(), 401)
			return "", "", err
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		tokenBody, authErr = s.doRequest(req)

		if _, ok := authBodyItem["password"]; ok {
			authBodyItem["password"] = []string{"xxx"}
		}

		if _, ok := authBodyItem["client_secret"]; ok {
			authBodyItem["client_secret"] = []string{"xxx"}
		}

		if authErr != nil {
			logger.Printf("Failed auth attempt %s", authBodyItem)
		}

		if authErr == nil {
			logger.Printf("Successful auth %s", authBodyItem)

			if _, ok := authBodyItem["username"]; ok {
				authEntity = authBodyItem["username"][0]
			} else {
				authEntity = authBodyItem["client_id"][0]
			}

			break
		}
	}

	if authErr != nil {
		logger.Printf("Failed all auth attempts %s", authErr)
		errStr := fmt.Sprintf("%s", authErr)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		http.Error(w, inverr.Error(), 401)
		return "", "", authErr
	}

	token := &Token{}
	uerr := json.Unmarshal(tokenBody, token)

	if uerr != nil {
		logger.Println(uerr)
		http.Error(w, uerr.Error(), 500)
		return "", "", uerr
	}

	return token.Value, authEntity, nil
}

func (s *APIClient) createClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client Client) (err error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf(controller.Config.ClientsURI, controller.Config.IdpURL, controller.Config.IdpRealm)
	byteArr, err := json.Marshal(client)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		http.Error(w, err.Error(), 500)
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	_, err = s.doRequest(req)

	if err != nil {
		logger.Println(err)
		errStr := fmt.Sprintf("%s", err)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		http.Error(w, inverr.Error(), 500)
		return &inverr
	}

	return nil
}

// getClientId - method for getting idp client id info
func (s *APIClient) getClientID(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client ClientWithSecret) (clientID string, err error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf(controller.Config.ClientsURI, controller.Config.IdpURL, controller.Config.IdpRealm)
	byteArr, err := json.Marshal(client)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		http.Error(w, err.Error(), 500)
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	resp, err := s.doRequest(req)

	if err != nil {
		logger.Println(err)
		errStr := fmt.Sprintf("%s", err)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		http.Error(w, inverr.Error(), 500)
		return "", &inverr
	}

	clientStruct := &ClientID{}

	addRoot := fmt.Sprintf(`{"root": %s}`, string(resp))
	jq := gojsonq.New().FromString(addRoot)

	if jq.Error() != nil {
		msg := fmt.Sprintf("Parsing response to jq failed, %s", jq.Errors())
		log.Println(msg)
		http.Error(w, msg, 500)
		return "", jq.Error()
	}

	jq.From("root").Where("clientId", "=", client.ClientID).Only("id")

	logger.Printf("Client %s id is %s", client.ClientID, clientStruct.ID)

	return clientStruct.ID, nil
}

// getClient - method for getting idp client info
func (s *APIClient) getClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client Client) (clientOut *ClientOut, err error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf(controller.Config.ClientsURI, controller.Config.IdpURL, controller.Config.IdpRealm)
	byteArr, err := json.Marshal(client)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		http.Error(w, err.Error(), 500)
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	resp, err := s.doRequest(req)

	if err != nil {
		logger.Println(err)
		errStr := fmt.Sprintf("%s", err)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		http.Error(w, inverr.Error(), 500)
		return nil, &inverr
	}

	clientStruct := &ClientOut{}

	addRoot := fmt.Sprintf(`{"root": %s}`, string(resp))
	jq := gojsonq.New().FromString(addRoot)

	if jq.Error() != nil {
		msg := fmt.Sprintf("Parsing response to jq failed, %s", jq.Errors())
		log.Println(msg)
		http.Error(w, msg, 500)
		return nil, jq.Error()
	}

	data := jq.From("root").Where("clientId", "=", client.ClientID).First()
	mapstructure.Decode(data, clientStruct)

	logger.Printf("Client %s id is %s", client.ClientID, clientStruct.ID)

	return clientStruct, nil
}

func (s *APIClient) getClientSecret(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	clientUID string) (clientSecret string, err error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf(controller.Config.ClientSecretURI, controller.Config.IdpURL, controller.Config.IdpRealm, clientUID)
	byteArr := []byte("")
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		http.Error(w, err.Error(), 500)
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.doRequest(req)

	if err != nil {
		logger.Println(err)
		errStr := fmt.Sprintf("%s", err)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		http.Error(w, inverr.Error(), 500)
		return "", &inverr
	}

	clientSecretStruct := &ClientSecret{}

	err = json.Unmarshal(resp, clientSecretStruct)

	if err != nil {
		logger.Println(err)
		http.Error(w, err.Error(), 500)
		return "", err
	}

	return clientSecretStruct.Value, nil
}

func (s *APIClient) updateClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	client Client, clientUID string) (err error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf(controller.Config.ClientURI, controller.Config.IdpURL, controller.Config.IdpRealm, clientUID)
	byteArr, err := json.Marshal(client)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		http.Error(w, err.Error(), 500)
		return err
	}

	logger.Println(url)

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	_, err = s.doRequest(req)

	if err != nil {
		logger.Println(err)
		errStr := fmt.Sprintf("%s", err)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		http.Error(w, inverr.Error(), 500)
		return &inverr
	}

	return nil
}

func (s *APIClient) deleteClient(
	w http.ResponseWriter,
	controller *Controller,
	token string,
	clientUID string) (err error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf(controller.Config.ClientURI, controller.Config.IdpURL, controller.Config.IdpRealm, clientUID)
	byteArr := []byte("")
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		http.Error(w, err.Error(), 500)
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	_, err = s.doRequest(req)

	if err != nil {
		logger.Println(err)
		errStr := fmt.Sprintf("%s", err)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		http.Error(w, inverr.Error(), 500)
		return &inverr
	}

	return nil
}

func (s *APIClient) createUser(
	config *Config,
	token string,
	user *User) (err error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf(config.UsersURI, config.IdpURL, config.IdpRealm)
	byteArr, err := json.Marshal(user)

	if err != nil {
		logger.Println(err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		return err
	}

	log.Println(url)

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	_, err = s.doRequest(req)

	if err != nil {
		logger.Println(err)
		errStr := fmt.Sprintf("%s", err)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		return &inverr
	}

	return nil
}

// getUserID - method for getting idp user id (really it has uid form)
func (s *APIClient) getUserID(
	config *Config,
	token string,
	user *User) (userID string, err error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf(config.UsersURI, config.IdpURL, config.IdpRealm)
	byteArr, err := json.Marshal(user)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	resp, err := s.doRequest(req)

	if err != nil {
		logger.Println(err)
		errStr := fmt.Sprintf("%s", err)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		return "", &inverr
	}

	userIDStruct := &UserID{}

	addRoot := fmt.Sprintf(`{"root": %s}`, string(resp))
	jq := gojsonq.New().FromString(addRoot)

	if jq.Error() != nil {
		msg := fmt.Sprintf("Parsing response to jq failed, %s", jq.Errors())
		log.Println(msg)
		return "", jq.Error()
	}

	res := jq.From("root").Where("username", "=", user.Username).Only("id")

	if jq.Error() != nil {
		msg := fmt.Sprintf("Querying id failed, %s", jq.Errors())
		log.Println(msg)
		return "", jq.Error()
	}

	resMapIntf, ok := res.([]interface{})

	if !ok {
		msg := "Failed assertion to array"
		logger.Println(msg)
		return "", apierror.InternalServerError()
	}

	resMapValIntf, ok := resMapIntf[0].(map[string]interface{})

	if !ok {
		msg := "Failed assertion to map"
		logger.Println(msg)
		return "", apierror.InternalServerError()
	}

	resID, ok := resMapValIntf["id"].(string)

	if !ok {
		msg := "Failed assertion to string"
		logger.Println(msg)
		return "", apierror.InternalServerError()
	}

	logger.Printf("User %s id is %s", user.Username, resID)

	userIDStruct.ID = resID

	return userIDStruct.ID, nil
}

func (s *APIClient) deleteUser(
	config *Config,
	token string,
	userUID string) (err error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf(config.UserURI, config.IdpURL, config.IdpRealm, userUID)
	byteArr := []byte("")
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	_, err = s.doRequest(req)

	if err != nil {
		logger.Println(err)
		errStr := fmt.Sprintf("%s", err)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		return &inverr
	}

	return nil
}

func (s *APIClient) setUserPassword(
	config *Config,
	token string,
	userCredential *UserSecret,
	userUID string) (err error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf(config.UserPasswordURI, config.IdpURL, config.IdpRealm, userUID)
	byteArr, err := json.Marshal(userCredential)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(byteArr))

	if err != nil {
		logger.Println(err)
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	_, err = s.doRequest(req)

	if err != nil {
		logger.Println(err)
		errStr := fmt.Sprintf("%s", err)
		inverr := apierror.ApiError{
			Code:    "10000",
			Message: errStr,
		}

		return &inverr
	}

	return nil
}
