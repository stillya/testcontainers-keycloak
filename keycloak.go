package keycloak

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"path/filepath"
)

const (
	defaultKeycloakImage         = "quay.io/keycloak/keycloak:20.0"
	defaultRealmImport           = "/opt/keycloak/data/import/"
	defaultKeycloakAdminUsername = "admin"
	defaultKeycloakAdminPassword = "admin"
	defaultKeycloakContextPath   = "/"
	keycloakAdminUsernameEnv     = "KEYCLOAK_ADMIN"
	keycloakAdminPasswordEnv     = "KEYCLOAK_ADMIN_PASSWORD"
	keycloakContextPathEnv       = "KEYCLOAK_CONTEXT_PATH"
	keycloakStartupCommand       = "start-dev"
)

// KeycloakContainer is a wrapper around testcontainers.Container
// that provides some convenience methods for working with Keycloak.
type KeycloakContainer struct {
	testcontainers.Container

	username    string
	password    string
	contextPath string
}

// GetAdminClient returns an AdminClient for the KeycloakContainer.
func (k *KeycloakContainer) GetAdminClient(ctx context.Context) (*AdminClient, error) {
	authServerURL, err := k.GetAuthServerURL(ctx)
	if err != nil {
		return nil, err
	}
	return NewAdminClient(&ctx, authServerURL, k.username, k.password)
}

// GetAuthServerURL returns the URL of the KeycloakContainer.
func (k *KeycloakContainer) GetAuthServerURL(ctx context.Context) (string, error) {
	host, err := k.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := k.MappedPort(ctx, "8080")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s%s", host, port.Port(), k.contextPath), nil
}

// RunContainer starts a new KeycloakContainer with the given options.
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*KeycloakContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: defaultKeycloakImage,
		Env: map[string]string{
			keycloakAdminUsernameEnv: defaultKeycloakAdminUsername,
			keycloakAdminPasswordEnv: defaultKeycloakAdminPassword,
		},
		ExposedPorts: []string{"8080/tcp"},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &KeycloakContainer{
		Container:   container,
		username:    genericContainerReq.Env[keycloakAdminUsernameEnv],
		password:    genericContainerReq.Env[keycloakAdminPasswordEnv],
		contextPath: genericContainerReq.Env[keycloakContextPathEnv],
	}, nil
}

// WithRealmImportFile is option to import a realm file into KeycloakContainer.
func WithRealmImportFile(realmImportFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		absPath, err := filepath.Abs(filepath.Dir(realmImportFile))
		if err != nil {
			return
		}
		// We have to mount because go-testcontainers does not support copying files to the container when target directory does not exist yet.
		// See this issue: https://github.com/testcontainers/testcontainers-go/issues/1336
		importFile := testcontainers.ContainerMount{
			Source: testcontainers.GenericBindMountSource{
				HostPath: absPath,
			},
			Target: defaultRealmImport,
		}
		req.Mounts = append(req.Mounts, importFile)

		processKeycloakArgs(req, []string{"--import-realm"})
	}
}

// WithAdminUsername is option to set the admin username for KeycloakContainer.
func WithAdminUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		if username == "" {
			username = defaultKeycloakAdminUsername
		}
		req.Env[keycloakAdminUsernameEnv] = username
	}
}

// WithAdminPassword is option to set the admin password for KeycloakContainer.
func WithAdminPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		if password == "" {
			password = defaultKeycloakAdminPassword
		}
		req.Env[keycloakAdminPasswordEnv] = password
	}
}

// WithContextPath is option to set the context path for KeycloakContainer.
func WithContextPath(contextPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		if contextPath == "" {
			contextPath = defaultKeycloakContextPath
		}
		req.Env[keycloakContextPathEnv] = contextPath
		processKeycloakArgs(req, []string{"--http-relative-path=" + contextPath})
	}
}

func processKeycloakArgs(req *testcontainers.GenericContainerRequest, args []string) {
	if len(req.Cmd) == 0 {
		req.Cmd = append([]string{keycloakStartupCommand}, args...)
		return
	}

	if req.Cmd[0] == keycloakStartupCommand {
		req.Cmd = append(req.Cmd, args...)
	} else if req.Cmd[0] != keycloakStartupCommand {
		req.Cmd = append([]string{keycloakStartupCommand}, req.Cmd...)
		req.Cmd = append(req.Cmd, args...)
	}
}
