package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/p53/idp-api/logging"
)

// App - main app structure
type App struct {
	router *mux.Router
}

// Config - structure holding configuration items for app
type Config struct {
	IdpURL          string
	ClientID        string
	ClientSecret    string
	ApiClientID     string
	ApiClientSecret string
	IdpAdmin        string
	IdpPass         string
	IdpRealm        string
	HTTPClient      APIClientIntf
	CheckURI        string
	ClientsURI      string
	ClientURI       string
	ClientSecretURI string
	TokenURI        string
	UsersURI        string
	UserURI         string
	UserPasswordURI string
}

// CreateApp - function for creating and initializing app
func CreateApp() *App {
	logger := logging.GetLogger()
	logger.Print("Starting...")

	apiClient := &APIClient{BaseClient: &http.Client{}}

	config := &Config{
		IdpURL:          os.Getenv("IDP_URL"),
		ClientID:        os.Getenv("CLIENT_ID"),
		ClientSecret:    os.Getenv("CLIENT_SECRET"),
		ApiClientID:     os.Getenv("API_CLIENT_ID"),
		ApiClientSecret: os.Getenv("API_CLIENT_SECRET"),
		IdpAdmin:        os.Getenv("IDP_ADMIN_USER"),
		IdpPass:         os.Getenv("IDP_ADMIN_PASSWORD"),
		IdpRealm:        os.Getenv("IDP_REALM"),
		HTTPClient:      apiClient,
		CheckURI:        "%s/auth/admin",
		ClientsURI:      "%s/auth/admin/realms/%s/clients",
		ClientURI:       "%s/auth/admin/realms/%s/clients/%s",
		ClientSecretURI: "%s/auth/admin/realms/%s/clients/%s/client-secret",
		TokenURI:        "%s/auth/realms/%s/protocol/openid-connect/token",
		UsersURI:        "%s/auth/admin/realms/%s/users",
		UserURI:         "%s/auth/admin/realms/%s/users/%s",
		UserPasswordURI: "%s/auth/admin/realms/%s/users/%s/reset-password",
	}

	controller := &Controller{Config: config}

	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()

	s.HandleFunc("/client", controller.DeleteResource).Methods("DELETE")
	s.HandleFunc("/client", controller.CreateResource).Methods("POST")
	s.HandleFunc("/client", controller.UpdateResource).Methods("PUT")

	r.HandleFunc("/health", controller.HealthCheck).Methods("GET")

	r.HandleFunc("/swagger.yml", controller.ReadSwagger).Methods("GET")

	http.Handle("/", s)

	app := &App{
		router: r,
	}

	return app
}

func (app *App) run() {
	logger := logging.GetLogger()

	srv := &http.Server{
		Handler: app.router,
		Addr:    "0.0.0.0:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(srv.ListenAndServe())
	logger.Print("Hello, log file!")
}
