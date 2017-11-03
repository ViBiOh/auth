# auth

[![Build Status](https://travis-ci.org/ViBiOh/auth.svg?branch=master)](https://travis-ci.org/ViBiOh/auth)
[![codecov](https://codecov.io/gh/ViBiOh/auth/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/auth)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/auth)](https://goreportcard.com/report/github.com/ViBiOh/auth)

Authentification for apps in microservices.

# Getting Started

You can use GitHub OAuth Provider or a simple username/password file for authentication.

## GitHub OAuth Provider

Create your OAuth app on [GitHub interface](https://github.com/settings/developers). The authorization callback URL must be set for pointing your app. The OAuth State is a random string use for verification by OAuth Provider, [see manual](https://developer.github.com/apps/building-integrations/setting-up-and-registering-oauth-apps/about-authorization-options-for-oauth-apps/).

## Basic Username/Password

Write user's credentials with the following format :

```
[id]:[username]:[bcrypt password],[id2]:[username2]:[bcrypt password2]
```

You can generate bcrypted password using `bin/bcrypt_pass`.

## Roles

You can configure roles by setting `-users` from CLI with the following format:

```
[user1]:[role1]|[role2],[user2]:[role1]
```

Username must match with the authentification providers (see previous sections).

## Build

In order to build the server stuff, run the following command.

```
make
```

It will compile both auth API server and password encrypter.

```
Usage of auth:
  -basicUsers string
        Basic users in the form "username:password,username2:password"
  -c string
        URL to healthcheck (check and exit)
  -corsHeaders string
        Access-Control-Allow-Headers (default "Content-Type")
  -corsMethods string
        Access-Control-Allow-Methods (default "GET")
  -corsOrigin string
        Access-Control-Allow-Origin (default "*")
  -csp string
        Content-Security-Policy (default "default-src 'self'")
  -githubClientId string
        GitHub OAuth Client ID
  -githubClientSecret string
        GitHub OAuth Client Secret
  -githubState string
        GitHub OAuth State
  -hsts
        Indicate Strict Transport Security (default true)
  -port string
        Listen port (default "1080")
  -prometheusMetricsHost string
        Prometheus - Allowed hostname to call metrics endpoint (default "localhost")
  -prometheusMetricsPath string
        Prometheus - Metrics endpoint path (default "/metrics")
  -rateCount int
        Rate IP limit (default 5000)
  -tls
        Serve TLS content (default true)
  -tlscert string
        TLS PEM Certificate file
  -tlshosts string
        TLS Self-signed certificate hosts, comma separated (default "localhost")
  -tlskey string
        TLS PEM Key file
```

Password encrypter accepts one argument, the password, and output the bcrypted one.
