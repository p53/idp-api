# IDP-API

## Development

For quick local development (running unit and integration tests) you can use provided docker-compose.yml file like this:

1. Run keycloak:

  ```
  docker-compose up
  ```

2. Inspect your newly created container ip

  ```
  docker inspect idpapi_keycloak-server_1 -f {{.NetworkSettings.Networks.idpapi_default.IPAddress}}
  ```

3. Edit controller_test.go file and getFuncTestConfig and set IP
   from above command, now run tests

For running all tests in docker environment just run:

  ```
  docker-compose -f docker-compose-test.yml up --exit-code-from idp-api-test --abort-on-container-exit

  ```

## Configuration

  Application is configurable through env vars:

  IDP_URL - url of IDP server

  CLIENT_ID - name of client used by app to verify users (must use resource owner credentials grant type)

  CLIENT_SECRET - secret for above client

  API_CLIENT_ID - name of IDP admin client
  
  API_CLIENT_SECRET - secret for above client

  IDP_ADMIN_USER - IDP admin user

  IDP_ADMIN_PASSWORD - IDP admin user password

  IDP_REALM - managed realm

## Usage

  Check swagger spec in swagger.yml in source code

  Creating client:

  ```
  curl -X POST -H 'Authorization: Basic <base64 encoded username:pass>' -d '{"clientId": "myclient"}' http://example.org/api/v1/client
  ```

  Updating client:

  ```
  curl -X PUT -H 'Authorization: Basic <base64 encoded username:pass>' -d '{"clientId": "myclient", "clientSecret": "somesecret"}' http://example.org/api/v1/client
  ```

  Deleting client:

  ```
  curl -X DELETE -H 'Authorization: Basic <base64 encoded username:pass>' -d '{"clientId": "myclient", "clientSecret": "somesecret"}' http://example.org/api/v1/client
  ```
