# auth

[![Build Status](https://travis-ci.org/ViBiOh/auth.svg?branch=master)](https://travis-ci.org/ViBiOh/auth)
[![codecov](https://codecov.io/gh/ViBiOh/auth/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/auth)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/auth)](https://goreportcard.com/report/github.com/ViBiOh/auth)

Authentification for apps in microservices.

# Getting Started

You can use GitHub OAuth Provider or a simple username/password file for authentication.

## GitHub OAuth Provider

Create your OAuth app on [GitHub interface](https://github.com/settings/developers). The authorization callback URL must be set for pointing your app.

### Username/Password file

Write user's credentials file with one line per user, each line having the following format :

```
[username],[bcrypt password]
```

You can generate bcrypted password using `bin/bcrypt_pass`.

You have ton configure `-basicFile` filepath on the Auth API server and change variable `BASIC_AUTH_ENABLED` value from `false` to `true` on front server.

## Build

In order to build the server stuff, run the following command.

```
make
```

It will compile both auth API server and password encrypter.

```
Usage of auth:
  -basicFile string
      Path of authentification file
  -c string
      URL to healthcheck (check and exit)
  -githubClientId string
      GitHub OAuth Client ID
  -githubClientSecret string
      GitHub OAuth Client Secret
  -githubState string
      GitHub OAuth State
  -port string
      Listen port (default "1080")
```

Password encrypter accepts one argument, the password, and output the bcrypted one.
