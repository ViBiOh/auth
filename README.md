# auth

[![Build Status](https://travis-ci.org/ViBiOh/auth.svg?branch=master)](https://travis-ci.org/ViBiOh/auth)
[![codecov](https://codecov.io/gh/ViBiOh/auth/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/auth)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/auth)](https://goreportcard.com/report/github.com/ViBiOh/auth)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=ViBiOh/auth)](https://dependabot.com)

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
  -address string
        [http] Listen address {AUTH_ADDRESS}
  -authRedirect string
        [auth] Redirect URL on Auth Success {AUTH_AUTH_REDIRECT}
  -basicUsers id:username:password,id2:username2:password2
        [basic] Users in the form id:username:password,id2:username2:password2 {AUTH_BASIC_USERS}
  -cert string
        [http] Certificate file {AUTH_CERT}
  -cookieDomain string
        [auth] Cookie Domain to Store Authentification {AUTH_COOKIE_DOMAIN}
  -corsCredentials
        [cors] Access-Control-Allow-Credentials {AUTH_CORS_CREDENTIALS}
  -corsExpose string
        [cors] Access-Control-Expose-Headers {AUTH_CORS_EXPOSE}
  -corsHeaders string
        [cors] Access-Control-Allow-Headers {AUTH_CORS_HEADERS} (default "Content-Type")
  -corsMethods string
        [cors] Access-Control-Allow-Methods {AUTH_CORS_METHODS} (default "GET")
  -corsOrigin string
        [cors] Access-Control-Allow-Origin {AUTH_CORS_ORIGIN} (default "*")
  -csp string
        [owasp] Content-Security-Policy {AUTH_CSP} (default "default-src 'self'; base-uri 'self'")
  -frameOptions string
        [owasp] X-Frame-Options {AUTH_FRAME_OPTIONS} (default "deny")
  -githubClientId string
        [github] OAuth Client ID {AUTH_GITHUB_CLIENT_ID}
  -githubClientSecret string
        [github] OAuth Client Secret {AUTH_GITHUB_CLIENT_SECRET}
  -githubScopes string
        [github] OAuth Scopes, comma separated {AUTH_GITHUB_SCOPES}
  -hsts
        [owasp] Indicate Strict Transport Security {AUTH_HSTS} (default true)
  -key string
        [http] Key file {AUTH_KEY}
  -port int
        [http] Listen port {AUTH_PORT} (default 1080)
  -prometheusPath string
        [prometheus] Path for exposing metrics {AUTH_PROMETHEUS_PATH} (default "/metrics")
  -url string
        [alcotest] URL to check {AUTH_URL}
  -userAgent string
        [alcotest] User-Agent for check {AUTH_USER_AGENT} (default "Golang alcotest")
```

Password encrypter accepts one argument, the password, and output the bcrypted one.
