# auth

[![Build](https://github.com/ViBiOh/auth/workflows/Build/badge.svg)](https://github.com/ViBiOh/auth/actions)

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
