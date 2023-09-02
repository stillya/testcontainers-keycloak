# Keycloak Testcontainer - [Testcontainers](https://www.testcontainers.org/) implementation for [Keycloak](https://www.keycloak.org/) SSO.

[![Build Status](https://github.com/stillya/testcontainers-keycloak/actions/workflows/go.yml/badge.svg)](https://github.com/stillya/testcontainers-keycloak/actions/workflows/go.yml)
[![Coverage](https://coveralls.io/repos/github/stillya/testcontainers-keycloak/badge.svg?branch=master)](https://coveralls.io/github/stillya/testcontainers-keycloak?branch=master)

* Native integration with [Testcontainers](https://www.testcontainers.org/).
* Customization via `realm.json` to create custom realms, users, clients, etc.
* Provides `AdminClient` to interact with Keycloak API.

## Installation

```bash
go get github.com/stillya/testcontainers-keycloak
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	keycloak "github.com/stillya/testcontainers-keycloak"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"testing"
)

var keycloakContainer *keycloak.KeycloakContainer

func Test_Example(t *testing.T) {
	ctx := context.Background()

	authServerURL, err := keycloakContainer.GetAuthServerURL(ctx)
	if err != nil {
		t.Errorf("GetAuthServerURL() error = %v", err)
		return
	}

	fmt.Println(authServerURL)
	// Output:
	// http://localhost:32768/auth
}

func TestMain(m *testing.M) {
	defer func() {
		if r := recover(); r != nil {
			shutDown()
			fmt.Println("Panic")
		}
	}()
	setup()
	code := m.Run()
	shutDown()
	os.Exit(code)
}

func setup() {
	var err error
	ctx := context.Background()
	keycloakContainer, err = RunContainer(ctx)
	if err != nil {
		panic(err)
	}
}

func shutDown() {
	ctx := context.Background()
	err := keycloakContainer.Terminate(ctx)
	if err != nil {
		panic(err)
	}
}

func RunContainer(ctx context.Context) (*keycloak.KeycloakContainer, error) {
	return keycloak.RunContainer(ctx,
		testcontainers.WithWaitStrategy(wait.ForListeningPort("8080/tcp")),
		keycloak.WithContextPath("/auth"),
		keycloak.WithRealmImportFile("../testdata/realm-export.json"),
		keycloak.WithAdminUsername("admin"),
		keycloak.WithAdminPassword("admin"),
	)
}
```
