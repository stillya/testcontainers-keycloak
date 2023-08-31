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

type KeycloakContainer struct {
	testcontainers.Container

	username    string
	password    string
	contextPath string
}

func (k *KeycloakContainer) GetAdminClient(ctx context.Context) (*AdminClient, error) {
	authServerURL, err := k.GetAuthServerURL(ctx)
	if err != nil {
		return nil, err
	}
	return NewAdminClient(&ctx, authServerURL, k.username, k.password)
}

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

func WithRealmImportFile(realmImportFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		absPath, err := filepath.Abs(filepath.Dir(realmImportFile))
		if err != nil {
			return
		}
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

func WithAdminUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		if username == "" {
			username = defaultKeycloakAdminUsername
		}
		req.Env[keycloakAdminUsernameEnv] = username
	}
}

func WithAdminPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		if password == "" {
			password = defaultKeycloakAdminPassword
		}
		req.Env[keycloakAdminPasswordEnv] = password
	}
}

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
