# auth

[![Build Status](https://travis-ci.org/ViBiOh/auth.svg?branch=master)](https://travis-ci.org/ViBiOh/auth)
[![codecov](https://codecov.io/gh/ViBiOh/auth/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/auth)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/auth)](https://goreportcard.com/report/github.com/ViBiOh/auth)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=ViBiOh/auth)](https://dependabot.com)

Authentification for apps in microservices.

# Getting Started

You can use a simple username/password file for authentication.

## Basic Username/Password

Write user's credentials with the following format :

```bash
[id]:[username]:[bcrypt password],[id2]:[username2]:[bcrypt password2]
```

You can generate bcrypted password using `go run cmd/bcrypt/bcrypt.go "password"`.

## Build

In order to build the whole stuff, run the following command.

```bash
make
```
Password encrypter accepts one argument, the password, and output the bcrypted one.
