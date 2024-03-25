package keycloak

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"path/filepath"
)

const (
	defaultKeycloakImage         = "quay.io/keycloak/keycloak:24.0"
	defaultRealmImport           = "/opt/keycloak/data/import/"
	defaultProviders             = "/opt/keycloak/providers/"
	tlsFilePath                  = "/opt/keycloak/conf"
	defaultKeycloakAdminUsername = "admin"
	defaultKeycloakAdminPassword = "admin"
	defaultKeycloakContextPath   = "/"
	keycloakAdminUsernameEnv     = "KEYCLOAK_ADMIN"
	keycloakAdminPasswordEnv     = "KEYCLOAK_ADMIN_PASSWORD"
	keycloakContextPathEnv       = "KEYCLOAK_CONTEXT_PATH"
	keycloakTlsEnv               = "KEYCLOAK_TLS"
	keycloakStartupCommand       = "start-dev"
	keycloakPort                 = "8080/tcp"
	keycloakHttpsPort            = "8443/tcp"
)

// KeycloakContainer is a wrapper around testcontainers.Container
// that provides some convenience methods for working with Keycloak.
type KeycloakContainer struct {
	testcontainers.Container

	username    string
	password    string
	enableTLS   bool
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
	if k.enableTLS {
		port, err := k.MappedPort(ctx, keycloakHttpsPort)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("https://%s:%s%s", host, port.Port(), k.contextPath), nil
	} else {
		port, err := k.MappedPort(ctx, keycloakPort)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("http://%s:%s%s", host, port.Port(), k.contextPath), nil
	}
}

// RunContainer starts a new KeycloakContainer with the given options.
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*KeycloakContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: defaultKeycloakImage,
		Env: map[string]string{
			keycloakAdminUsernameEnv: defaultKeycloakAdminUsername,
			keycloakAdminPasswordEnv: defaultKeycloakAdminPassword,
		},
		ExposedPorts: []string{keycloakPort},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	if genericContainerReq.WaitingFor == nil {
		contextPath := genericContainerReq.Env[keycloakContextPathEnv]
		if contextPath == "" {
			contextPath = defaultKeycloakContextPath
		}
		if genericContainerReq.Env[keycloakTlsEnv] != "" {
			genericContainerReq.WaitingFor = wait.ForAll(wait.ForHTTP(contextPath).
				WithPort(keycloakHttpsPort).
				WithTLS(true).
				WithAllowInsecure(true),
				wait.ForLog("Running the server"))
		} else {
			genericContainerReq.WaitingFor = wait.ForAll(wait.ForHTTP(contextPath),
				wait.ForLog("Running the server"))
		}
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
		enableTLS:   genericContainerReq.Env[keycloakTlsEnv] != "",
	}, nil
}

// WithRealmImportFile is option to import a realm file into KeycloakContainer.
func WithRealmImportFile(realmImportFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		realmFile := testcontainers.ContainerFile{
			HostFilePath:      realmImportFile,
			ContainerFilePath: defaultRealmImport + filepath.Base(realmImportFile),
			FileMode:          0o755,
		}
		req.Files = append(req.Files, realmFile)

		processKeycloakArgs(req, []string{"--import-realm"})
	}
}

// WithProviders is option to set the providers for KeycloakContainer.
// Providers should be packaged ina Java Archive (JAR) file.
// See https://www.keycloak.org/server/configuration-provider
func WithProviders(providerFiles ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		for _, providerFile := range providerFiles {
			provider := testcontainers.ContainerFile{
				HostFilePath:      providerFile,
				ContainerFilePath: defaultProviders + filepath.Base(providerFile),
				FileMode:          0o755,
			}
			req.Files = append(req.Files, provider)
		}
	}
}

// WithTLS is option to enable TLS for KeycloakContainer.
func WithTLS(certFile, keyFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.ExposedPorts = []string{keycloakHttpsPort}
		cf := testcontainers.ContainerFile{
			HostFilePath:      certFile,
			ContainerFilePath: tlsFilePath + "/tls.crt",
			FileMode:          0o755,
		}
		kf := testcontainers.ContainerFile{
			HostFilePath:      keyFile,
			ContainerFilePath: tlsFilePath + "/tls.key",
			FileMode:          0o755,
		}

		req.Files = append(req.Files, cf, kf)

		req.Env[keycloakTlsEnv] = "true"
		processKeycloakArgs(req,
			[]string{"--https-certificate-file=" + tlsFilePath + "/tls.crt",
				"--https-certificate-key-file=" + tlsFilePath + "/tls.key"},
		)
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
