# auth

[![Build Status](https://travis-ci.com/ViBiOh/auth.svg?branch=master)](https://travis-ci.com/ViBiOh/auth)
[![codecov](https://codecov.io/gh/ViBiOh/auth/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/auth)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/auth/v2)](https://goreportcard.com/report/github.com/ViBiOh/auth/v2)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=ViBiOh/auth)](https://dependabot.com)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=ViBiOh_auth&metric=alert_status)](https://sonarcloud.io/dashboard?id=ViBiOh_auth)

Authentification for apps in microservices.

# Getting Started

You can use a simple login/password file for authentication.

## Basic Login/Password

Write user's credentials with the following format :

```bash
[id]:[login]:[bcrypt password],[id2]:[login2]:[bcrypt password2]
```

You can generate bcrypted password using `go run cmd/bcrypt/bcrypt.go "password"`.

## Build

In order to build the whole stuff, run the following command.

```bash
make
```
Password encrypter accepts one argument, the password, and output the bcrypted one.
