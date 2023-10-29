package keycloak

import (
	"context"
	"crypto/tls"
	"github.com/testcontainers/testcontainers-go"
	"net/http"
	"testing"
)

const (
	username = "testUsername"
	password = "testPassword"
	realm    = "Test"
	client   = "test-app"
)

func TestKeycloak(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		image  string
		useTLS bool
		option testcontainers.CustomizeRequestOption
	}{
		{
			name:   "KeycloakV20CustomOption",
			image:  "quay.io/keycloak/keycloak:20.0",
			option: WithCustomOption(),
		},
		{
			name:   "KeycloakV20WithTLS",
			image:  "quay.io/keycloak/keycloak:20.0",
			option: WithTLS("testdata/tls.crt", "testdata/tls.key"),
			useTLS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := RunContainer(ctx,
				tt.option,
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

			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}
			oidConfResp, err := client.Get(authServerURL + "/realms/" + realm + "/.well-known/openid-configuration")
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
		name   string
		image  string
		useTLS bool
		option testcontainers.CustomizeRequestOption
	}{
		{
			name:   "KeycloakV20",
			image:  "quay.io/keycloak/keycloak:20.0",
			useTLS: false,
			option: WithCustomOption(),
		},
		{
			name:   "KeycloakV20WithTLS",
			image:  "quay.io/keycloak/keycloak:20.0",
			option: WithTLS("testdata/tls.crt", "testdata/tls.key"),
			useTLS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := RunContainer(ctx,
				tt.option,
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

func TestKeycloakContainer_GetAuthServerURL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		image  string
		useTLS bool
		option testcontainers.CustomizeRequestOption
	}{
		{
			name:   "KeycloakV20",
			image:  "quay.io/keycloak/keycloak:20.0",
			option: WithCustomOption(),
		},
		{
			name:   "KeycloakV20WithTLS",
			image:  "quay.io/keycloak/keycloak:20.0",
			option: WithTLS("testdata/tls.crt", "testdata/tls.key"),
			useTLS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := RunContainer(ctx,
				tt.option,
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

			if tt.useTLS {
				port, err := container.MappedPort(ctx, keycloakHttpsPort)
				if authServerURL != "https://localhost:"+port.Port()+"/auth" {
					t.Errorf("GetAuthServerURL() error = %v", err)
					return
				}
				return
			} else {
				port, err := container.MappedPort(ctx, keycloakPort)
				if authServerURL != "http://localhost:"+port.Port()+"/auth" {
					t.Errorf("GetAuthServerURL() error = %v", err)
					return
				}
			}
		})
	}
}

func WithCustomOption() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Cmd = []string{"--health-enabled=false"}
	}
}
