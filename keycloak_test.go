package keycloak

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"testing"
)

const (
	username = "testUsername"
	password = "testPassword"
	realm    = "Test"
	client   = "test-app"
)

func TestPostgres(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		image string
	}{
		{
			name:  "KeycloakV20",
			image: "quay.io/keycloak/keycloak:20.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := RunContainer(ctx,
				testcontainers.WithWaitStrategy(wait.ForListeningPort("8080/tcp")),
				WithContextPath("/auth"),
				WithRealmImportFile("testdata/realm-export.json"),
				WithAdminUsername(username),
				WithAdminPassword(password),
			)
			if err != nil {
				t.Errorf("RunContainer() error = %v", err)
				return
			}

			t.Cleanup(func() {
				err := container.Terminate(ctx)
				if err != nil {
					t.Errorf("Terminate() error = %v", err)
					return
				}
			})

			authServerURL, err := container.GetAuthServerURL(ctx)
			if err != nil {
				t.Errorf("GetAuthServerURL() error = %v", err)
				return
			}

			oidConfResp, err := http.Get(authServerURL + "/realms/" + realm + "/.well-known/openid-configuration")
			if err != nil {
				t.Errorf("http.Get() error = %v", err)
				return
			}

			if oidConfResp.StatusCode != http.StatusOK {
				t.Errorf("http.Get() error = %v", err)
				return
			}
		})
	}
}

func TestKeycloakContainer_GetAdminClient(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		image string
	}{
		{
			name:  "KeycloakV20",
			image: "quay.io/keycloak/keycloak:20.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := RunContainer(ctx,
				testcontainers.WithWaitStrategy(wait.ForListeningPort("8080/tcp")),
				WithContextPath("/auth"),
				WithRealmImportFile("testdata/realm-export.json"),
				WithAdminUsername(username),
				WithAdminPassword(password),
			)
			if err != nil {
				t.Errorf("RunContainer() error = %v", err)
				return
			}

			t.Cleanup(func() {
				err := container.Terminate(ctx)
				if err != nil {
					t.Errorf("Terminate() error = %v", err)
					return
				}
			})

			adminClient, err := container.GetAdminClient(ctx)
			if err != nil {
				t.Errorf("GetAdminClient() error = %v", err)
				return
			}

			c, err := adminClient.GetClient(realm, client)
			if err != nil {
				t.Errorf("GetClient() error = %v", err)
				return
			}

			if *c.ClientID != client {
				t.Errorf("GetClient() error = %v", err)
				return
			}
		})
	}
}
