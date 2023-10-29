package _example

import (
	"context"
	"fmt"
	keycloak "github.com/stillya/testcontainers-keycloak"
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

func Test_AnotherExample(t *testing.T) {
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
		keycloak.WithContextPath("/auth"),
		keycloak.WithRealmImportFile("../testdata/realm-export.json"),
		keycloak.WithAdminUsername("admin"),
		keycloak.WithAdminPassword("admin"),
	)
}
