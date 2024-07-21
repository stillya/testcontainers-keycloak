package keycloak

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"path/filepath"
)

const (
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

// Run starts a new KeycloakContainer with the given options.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*KeycloakContainer, error) {
	req := testcontainers.ContainerRequest{
		// TODO: Add custom container registry substitutor when this feature will be included in new testcontainers-go release
		// https://github.com/testcontainers/testcontainers-go/pull/2647
		Image: img,
		Env: map[string]string{
			keycloakAdminUsernameEnv: defaultKeycloakAdminUsername,
			keycloakAdminPasswordEnv: defaultKeycloakAdminPassword,
		},
		ExposedPorts: []string{keycloakPort},
		Cmd:          []string{keycloakStartupCommand},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
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
	return func(req *testcontainers.GenericContainerRequest) error {
		realmFile := testcontainers.ContainerFile{
			HostFilePath:      realmImportFile,
			ContainerFilePath: defaultRealmImport + filepath.Base(realmImportFile),
			FileMode:          0o755,
		}
		req.Files = append(req.Files, realmFile)

		processKeycloakArgs(req, []string{"--import-realm"})

		return nil
	}
}

// WithProviders is option to set the providers for KeycloakContainer.
// Providers should be packaged ina Java Archive (JAR) file.
// See https://www.keycloak.org/server/configuration-provider
func WithProviders(providerFiles ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		for _, providerFile := range providerFiles {
			provider := testcontainers.ContainerFile{
				HostFilePath:      providerFile,
				ContainerFilePath: defaultProviders + filepath.Base(providerFile),
				FileMode:          0o755,
			}
			req.Files = append(req.Files, provider)
		}

		return nil
	}
}

// WithTLS is option to enable TLS for KeycloakContainer.
func WithTLS(certFile, keyFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
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

		return nil
	}
}

// WithAdminUsername is option to set the admin username for KeycloakContainer.
func WithAdminUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if username == "" {
			username = defaultKeycloakAdminUsername
		}
		req.Env[keycloakAdminUsernameEnv] = username

		return nil
	}
}

// WithAdminPassword is option to set the admin password for KeycloakContainer.
func WithAdminPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if password == "" {
			password = defaultKeycloakAdminPassword
		}
		req.Env[keycloakAdminPasswordEnv] = password

		return nil
	}
}

// WithContextPath is option to set the context path for KeycloakContainer.
func WithContextPath(contextPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if contextPath == "" {
			contextPath = defaultKeycloakContextPath
		}
		req.Env[keycloakContextPathEnv] = contextPath
		processKeycloakArgs(req, []string{"--http-relative-path=" + contextPath})

		return nil
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
