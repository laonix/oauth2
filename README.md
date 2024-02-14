# OAuth2 Server

## Task
- Create a Golang http server that issues JWT Access Tokens (rfc7519) using Client Credentials Grant with Basic Authentication (rfc6749) in endpoint with path /token
- Return a sign  tokens with a self-made RS256 key
- Provide a deployment manifests to deploy it inside Kubernetes

## Dependencies
The [oauth2 module](https://github.com/go-oauth2/oauth2) was used to develop OAuth2 Server. From this module these packages were used:
- `store` for storing tokens and users in memory
- `manager` for working with tokens
- `server`. This package contains all the functions necessary for the server to operate. The server can be flexibly configured for various scenarios. This package is also compatible with `net/http`

## Assumptions
- Since the condition did not say that it was necessary to make an endpoint for adding users, therefore one user with `client_id: "client_id"` and `secret: "client_secret"` was hardcoded.
- For the http server, `net/http` was used
- `/secure` endpoint was created that verifies the token and returns status 200 if the token is valid.
- `viper` was used to configure the application. I also created a `config.yaml` file for local development. Configuration values can be overridden by environment variables. Currently there are these variables: `HTTP_PORT`, `HTTP_TIMEOUT`, `JWT_ACCESS_TOKEN_EXPIRES_IN` and `JWT_SECRET`. For a production environment, you definitely need to override `JWT_SECRET`.
- Since `memory` storage was used to store tokens, `replicas` in `deployment.yaml` was set to 1.
- The Postman collection was created to simplify testing.