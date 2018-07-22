# auth

[![Build Status](https://travis-ci.org/ViBiOh/auth.svg?branch=master)](https://travis-ci.org/ViBiOh/auth)
[![codecov](https://codecov.io/gh/ViBiOh/auth/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/auth)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/auth)](https://goreportcard.com/report/github.com/ViBiOh/auth)

Authentification for apps in microservices.

# Getting Started

You can use GitHub OAuth Provider or a simple username/password file for
authentication.

## GitHub OAuth Provider

Create your OAuth app on
[GitHub interface](https://github.com/settings/developers). The authorization
callback URL must be set for pointing your app. The OAuth State is a random
string use for verification by OAuth Provider,
[see manual](https://developer.github.com/apps/building-integrations/setting-up-and-registering-oauth-apps/about-authorization-options-for-oauth-apps/).

## Basic Username/Password

Write user's credentials with the following format :

```bash
[id]:[username]:[bcrypt password],[id2]:[username2]:[bcrypt password2]
```

You can generate bcrypted password using `go run bcrypt/bcrypt.go "password"`.

## Build

In order to build the server stuff, run the following command.

```bash
make
```

It will compile both auth API server and password encrypter.

```bash
Usage of auth:
  -authRedirect string
      [auth] Redirect URL on Auth Success
  -basicUsers string
      [Basic] Users in the form "id:username:password,id2:username2:password2"
  -cookieDomain string
      [auth] Cookie Domain to Store Authentification
  -corsCredentials
      [cors] Access-Control-Allow-Credentials
  -corsExpose string
      [cors] Access-Control-Expose-Headers
  -corsHeaders string
      [cors] Access-Control-Allow-Headers (default "Content-Type")
  -corsMethods string
      [cors] Access-Control-Allow-Methods (default "GET")
  -corsOrigin string
      [cors] Access-Control-Allow-Origin (default "*")
  -csp string
      [owasp] Content-Security-Policy (default "default-src 'self'; base-uri 'self'")
  -frameOptions string
      [owasp] X-Frame-Options (default "deny")
  -githubClientId string
      [GitHub] OAuth Client ID
  -githubClientSecret string
      [GitHub] OAuth Client Secret
  -githubScopes string
      [GitHub] OAuth Scopes, comma separated
  -hsts
      [owasp] Indicate Strict Transport Security (default true)
  -port int
      Listen port (default 1080)
  -rollbarEnv string
      [rollbar] Environment (default "prod")
  -rollbarServerRoot string
      [rollbar] Server Root
  -rollbarToken string
      [rollbar] Token
  -tls
      Serve TLS content (default true)
  -tlsCert string
      [tls] PEM Certificate file
  -tlsHosts string
      [tls] Self-signed certificate hosts, comma separated (default "localhost")
  -tlsKey string
      [tls] PEM Key file
  -tlsOrganization string
      [tls] Self-signed certificate organization (default "ViBiOh")
  -tracingAgent string
      [opentracing] Jaeger Agent (e.g. host:port) (default "jaeger:6831")
  -tracingName string
      [opentracing] Service name
  -twitterKey string
      [Twitter] Consumer Key
  -twitterSecret string
      [Twitter] Consumer Secret
  -url string
      [health] URL to check
  -userAgent string
      [health] User-Agent used (default "Golang alcotest")
```

Password encrypter accepts one argument, the password, and output the bcrypted one.
