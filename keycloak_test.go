package keycloak

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"slices"
	"testing"

	"github.com/testcontainers/testcontainers-go"
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
		envs   map[string]string
	}{
		{
			name:  "KeycloakV26Basic",
			image: "keycloak/keycloak:26.0",
		},
		{
			name:  "KeycloakV26BasicWithCustomHub",
			image: "keycloak/keycloak:26.0",
			envs: map[string]string{
				"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX": "quay.io",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envs {
				err := os.Setenv(k, v)
				if err != nil {
					t.Errorf("Setenv() error = %v", err)
				}
				t.Cleanup(func() {
					err := os.Unsetenv(k)
					if err != nil {
						t.Errorf("Unsetenv() error = %v", err)
					}
				})
			}

			container, err := Run(ctx,
				tt.image,
			)
			if err != nil {
				t.Errorf("Run() error = %v", err)
				return
			}

			testcontainers.CleanupContainer(t, container)
		})
	}
}

func TestKeycloakWithOptions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		image   string
		useTLS  bool
		options []testcontainers.ContainerCustomizer
	}{
		{
			name:  "KeycloakV24WithCustomOption",
			image: "keycloak/keycloak:24.0",
			options: []testcontainers.ContainerCustomizer{
				WithCustomOption(),
				WithContextPath("/auth"),
				WithRealmImportFile("testdata/realm-export.json"),
				WithAdminUsername(username),
				WithAdminPassword(password),
			},
		},
		{
			name:  "KeycloakV24WithTLS",
			image: "keycloak/keycloak:24.0",
			options: []testcontainers.ContainerCustomizer{
				WithTLS("testdata/tls.crt", "testdata/tls.key"),
				WithContextPath("/auth"),
				WithRealmImportFile("testdata/realm-export.json"),
				WithAdminUsername(username),
				WithAdminPassword(password),
			},
			useTLS: true,
		},
		{
			name:  "KeycloakV26WithCustomOption",
			image: "keycloak/keycloak:26.0",
			options: []testcontainers.ContainerCustomizer{
				WithCustomOption(),
				WithContextPath("/auth"),
				WithRealmImportFile("testdata/realm-export.json"),
				WithAdminUsername(username),
				WithAdminPassword(password),
			},
		},
		{
			name:  "KeycloakV26WithTLS",
			image: "keycloak/keycloak:26.0",
			options: []testcontainers.ContainerCustomizer{
				WithTLS("testdata/tls.crt", "testdata/tls.key"),
				WithContextPath("/auth"),
				WithRealmImportFile("testdata/realm-export.json"),
				WithAdminUsername(username),
				WithAdminPassword(password),
			},
			useTLS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := Run(ctx,
				tt.image,
				tt.options...,
			)
			if err != nil {
				t.Errorf("Run() error = %v", err)
				return
			}

			testcontainers.CleanupContainer(t, container)

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
		name    string
		image   string
		useTLS  bool
		options []testcontainers.ContainerCustomizer
	}{
		{
			name:   "KeycloakV24WithCustomOption",
			image:  "keycloak/keycloak:24.0",
			useTLS: false,
			options: []testcontainers.ContainerCustomizer{
				WithCustomOption(),
				WithContextPath("/auth"),
				WithRealmImportFile("testdata/realm-export.json"),
				WithAdminUsername(username),
				WithAdminPassword(password),
			},
		},
		{
			name:  "KeycloakV24WithTLS",
			image: "keycloak/keycloak:24.0",
			options: []testcontainers.ContainerCustomizer{
				WithTLS("testdata/tls.crt", "testdata/tls.key"),
				WithContextPath("/auth"),
				WithRealmImportFile("testdata/realm-export.json"),
				WithAdminUsername(username),
				WithAdminPassword(password),
			},
			useTLS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := Run(ctx,
				tt.image,
				tt.options...,
			)
			if err != nil {
				t.Errorf("Run() error = %v", err)
				return
			}

			testcontainers.CleanupContainer(t, container)

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
		name    string
		image   string
		useTLS  bool
		options []testcontainers.ContainerCustomizer
	}{
		{
			name:    "KeycloakV24Basic",
			image:   "keycloak/keycloak:24.0",
			options: []testcontainers.ContainerCustomizer{},
		},
		{
			name:  "KeycloakV24WithCustomOption",
			image: "keycloak/keycloak:24.0",
			options: []testcontainers.ContainerCustomizer{
				WithCustomOption(),
			},
		},
		{
			name:  "KeycloakV24WithTLS",
			image: "keycloak/keycloak:24.0",
			options: []testcontainers.ContainerCustomizer{
				WithTLS("testdata/tls.crt", "testdata/tls.key"),
			},
			useTLS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := Run(ctx,
				tt.image,
				tt.options...,
			)
			if err != nil {
				t.Errorf("Run() error = %v", err)
				return
			}

			testcontainers.CleanupContainer(t, container)

			authServerURL, err := container.GetAuthServerURL(ctx)
			if err != nil {
				t.Errorf("GetAuthServerURL() error = %v", err)
				return
			}

			if tt.useTLS {
				port, err := container.MappedPort(ctx, keycloakHttpsPort)
				if authServerURL != "https://localhost:"+port.Port()+container.contextPath {
					t.Errorf("GetAuthServerURL() error = %v", err)
					return
				}
				return
			} else {
				port, err := container.MappedPort(ctx, keycloakPort)
				if authServerURL != "http://localhost:"+port.Port()+container.contextPath {
					t.Errorf("GetAuthServerURL() error = %v", err)
					return
				}
			}
		})
	}
}

func TestProcessKeycloakArgs(t *testing.T) {
	tests := []struct {
		name string
		cmd  []string
		want []string
	}{
		{
			name: "EmptyCmd",
			cmd:  nil,
			want: []string{"start-dev", "--import-realm"},
		},
		{
			name: "DevMode",
			cmd:  []string{"start-dev"},
			want: []string{"start-dev", "--import-realm"},
		},
		{
			name: "ProductionMode",
			cmd:  []string{"start"},
			want: []string{"start", "--import-realm"},
		},
		{
			name: "ProductionModeWithOptimized",
			cmd:  []string{"start", "--optimized"},
			want: []string{"start", "--optimized", "--import-realm"},
		},
		{
			name: "GlobalOptionBeforeCommand",
			cmd:  []string{"--verbose", "start"},
			want: []string{"--verbose", "start", "--import-realm"},
		},
		{
			name: "OptionsWithoutCommand",
			cmd:  []string{"--health-enabled=false"},
			want: []string{"start-dev", "--health-enabled=false", "--import-realm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{Cmd: tt.cmd},
			}

			processKeycloakArgs(req, []string{"--import-realm"})

			if !slices.Equal(req.Cmd, tt.want) {
				t.Errorf("processKeycloakArgs() cmd = %v, want %v", req.Cmd, tt.want)
			}
		})
	}
}

func WithCustomOption() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Cmd = append(req.Cmd, "--health-enabled=false")

		return nil
	}
}
