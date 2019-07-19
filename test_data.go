package main

var testPayload = `{
  "clientId":"test",
  "publicClient": true,
  "directAccessGrantsEnabled": true,
  "serviceAccountsEnabled": false,
  "standardFlowEnabled": false,
  "implicitFlowEnabled": false
}`

var testBadPayload = `{
  "clientIdDDDDDDD":"test",
  "directAccessGrantsEnabled": true,
  "serviceAccountsEnabled": false,
  "standardFlowEnabled": false,
  "implicitFlowEnabled": false
}`

var testSecretPayload = `{
  "clientId":"test",
  "directAccessGrantsEnabled": true,
  "serviceAccountsEnabled": false,
  "standardFlowEnabled": false,
  "implicitFlowEnabled": false,
  "clientSecret": "testsecret"
}`

var testBadTokenBody = `{
  "access_token": test_access_token",
  "expires_in": 120,
  "refresh_expires_in": 1800,
  "refresh_token": "test_refresh_token",
  "token_type": "bearer",
  "not-before-policy": 0,
  "session_state": "3585e2c9-387b-42b2-a45d-de2aebae14cf",
  "scope": "profile email"
}`

var testTokenBody = `{
  "access_token": "test_access_token",
  "expires_in": 120,
  "refresh_expires_in": 1800,
  "refresh_token": "test_refresh_token",
  "token_type": "bearer",
  "not-before-policy": 0,
  "session_state": "3585e2c9-387b-42b2-a45d-de2aebae14cf",
  "scope": "profile email"
}`

var testClientsData = `[
      {
          "access": {
              "configure": true,
              "manage": true,
              "view": true
          },
          "attributes": {},
          "authenticationFlowBindingOverrides": {},
          "baseUrl": "/auth/admin/pan-net.eu/console/index.html",
          "bearerOnly": false,
          "clientAuthenticatorType": "client-secret",
          "clientId": "security-admin-console",
          "consentRequired": false,
          "defaultClientScopes": [
              "web-origins",
              "role_list",
              "roles",
              "profile",
              "email"
          ],
          "directAccessGrantsEnabled": false,
          "enabled": true,
          "frontchannelLogout": false,
          "fullScopeAllowed": false,
          "id": "8833abea-f37b-4c28-b353-63cafa3d5a26",
          "implicitFlowEnabled": false,
          "name": "${client_security-admin-console}",
          "nodeReRegistrationTimeout": 0,
          "notBefore": 0,
          "optionalClientScopes": [
              "address",
              "phone",
              "offline_access"
          ],
          "protocol": "openid-connect",
          "protocolMappers": [
              {
                  "config": {
                      "access.token.claim": "true",
                      "claim.name": "locale",
                      "id.token.claim": "true",
                      "jsonType.label": "String",
                      "user.attribute": "locale",
                      "userinfo.token.claim": "true"
                  },
                  "consentRequired": false,
                  "id": "6f17ccc5-a7b4-42d9-b402-c168c38b6101",
                  "name": "locale",
                  "protocol": "openid-connect",
                  "protocolMapper": "oidc-usermodel-attribute-mapper"
              }
          ],
          "publicClient": true,
          "redirectUris": [
              "/auth/admin/pan-net.eu/console/*"
          ],
          "serviceAccountsEnabled": false,
          "standardFlowEnabled": true,
          "surrogateAuthRequired": false,
          "webOrigins": []
      },
      {
          "access": {
              "configure": true,
              "manage": true,
              "view": true
          },
          "attributes": {},
          "authenticationFlowBindingOverrides": {},
          "bearerOnly": false,
          "clientAuthenticatorType": "client-secret",
          "clientId": "test",
          "consentRequired": false,
          "defaultClientScopes": [
              "web-origins",
              "pan-net.eu-client-template-openid-connect",
              "role_list",
              "roles",
              "profile",
              "pan-net.eu-client-template-saml",
              "email"
          ],
          "directAccessGrantsEnabled": false,
          "enabled": true,
          "frontchannelLogout": false,
          "fullScopeAllowed": true,
          "id": "40b5444c-5990-496d-bb67-64c535df8dc4",
          "implicitFlowEnabled": false,
          "nodeReRegistrationTimeout": -1,
          "notBefore": 0,
          "optionalClientScopes": [
              "address",
              "phone",
              "offline_access"
          ],
          "protocol": "openid-connect",
          "publicClient": false,
          "redirectUris": [],
          "serviceAccountsEnabled": false,
          "standardFlowEnabled": true,
          "surrogateAuthRequired": false,
          "webOrigins": []
      }
    ]
`

var testClientSecret = `{"type":"secret","value":"test_secret"}`

var testUsersData = `[
{
    "id": "06b4f835-a8c7-40f4-887b-cd76c7623267",
    "createdTimestamp": 1560863069395,
    "username": "test",
    "enabled": true,
    "totp": false,
    "emailVerified": false,
    "disableableCredentialTypes": [],
    "requiredActions": [],
    "notBefore": 0,
    "access": {
      "manageGroupMembership": true,
      "view": true,
      "mapRoles": true,
      "impersonate": true,
      "manage": true
    }
  }
]
`

var testUserPayload = `{
  "username": "test",
  "enabled": true
}
`

var testUserSecretPayload = `{
  "type": "password",
  "value": "test",
  "temporary": false
}
`

var testNewClient = `{
  "clientId":"test1",
  "directAccessGrantsEnabled": true,
  "serviceAccountsEnabled": false,
  "standardFlowEnabled": false,
  "implicitFlowEnabled": false
}`

var testUpdateClient = `{
  "clientId":"test1",
  "directAccessGrantsEnabled": true,
  "serviceAccountsEnabled": true,
  "standardFlowEnabled": true,
  "implicitFlowEnabled": false
}`

var testServiceClient = `{
  "clientId":"test_service",
  "directAccessGrantsEnabled": false,
  "serviceAccountsEnabled": true,
  "standardFlowEnabled": false,
  "implicitFlowEnabled": false
}`
