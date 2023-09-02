package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	adminClientID = "admin-cli"
	masterRealm   = "master"
)

// Token represents a Keycloak token.
type Token struct {
	AccessToken      string `json:"access_token"`
	IDToken          string `json:"id_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

// Client represents a Keycloak client(https://www.keycloak.org/docs-api/19.0.3/javadocs/org/keycloak/representations/idm/ClientRepresentation.html).
type Client struct {
	Access                             *map[string]interface{} `json:"access,omitempty"`
	AdminURL                           *string                 `json:"adminUrl,omitempty"`
	Attributes                         *map[string]string      `json:"attributes,omitempty"`
	AuthenticationFlowBindingOverrides *map[string]string      `json:"authenticationFlowBindingOverrides,omitempty"`
	AuthorizationServicesEnabled       *bool                   `json:"authorizationServicesEnabled,omitempty"`
	BaseURL                            *string                 `json:"baseUrl,omitempty"`
	BearerOnly                         *bool                   `json:"bearerOnly,omitempty"`
	ClientAuthenticatorType            *string                 `json:"clientAuthenticatorType,omitempty"`
	ClientID                           *string                 `json:"clientId,omitempty"`
	ConsentRequired                    *bool                   `json:"consentRequired,omitempty"`
	DefaultClientScopes                *[]string               `json:"defaultClientScopes,omitempty"`
	DefaultRoles                       *[]string               `json:"defaultRoles,omitempty"`
	Description                        *string                 `json:"description,omitempty"`
	DirectAccessGrantsEnabled          *bool                   `json:"directAccessGrantsEnabled,omitempty"`
	Enabled                            *bool                   `json:"enabled,omitempty"`
	FrontChannelLogout                 *bool                   `json:"frontchannelLogout,omitempty"`
	FullScopeAllowed                   *bool                   `json:"fullScopeAllowed,omitempty"`
	ID                                 *string                 `json:"id,omitempty"`
	ImplicitFlowEnabled                *bool                   `json:"implicitFlowEnabled,omitempty"`
	Name                               *string                 `json:"name,omitempty"`
	NodeReRegistrationTimeout          *int32                  `json:"nodeReRegistrationTimeout,omitempty"`
	NotBefore                          *int32                  `json:"notBefore,omitempty"`
	OptionalClientScopes               *[]string               `json:"optionalClientScopes,omitempty"`
	Origin                             *string                 `json:"origin,omitempty"`
	Protocol                           *string                 `json:"protocol,omitempty"`
	PublicClient                       *bool                   `json:"publicClient,omitempty"`
	RedirectURIs                       *[]string               `json:"redirectUris,omitempty"`
	RegisteredNodes                    *map[string]int         `json:"registeredNodes,omitempty"`
	RegistrationAccessToken            *string                 `json:"registrationAccessToken,omitempty"`
	RootURL                            *string                 `json:"rootUrl,omitempty"`
	Secret                             *string                 `json:"secret,omitempty"`
	ServiceAccountsEnabled             *bool                   `json:"serviceAccountsEnabled,omitempty"`
	StandardFlowEnabled                *bool                   `json:"standardFlowEnabled,omitempty"`
	SurrogateAuthRequired              *bool                   `json:"surrogateAuthRequired,omitempty"`
	WebOrigins                         *[]string               `json:"webOrigins,omitempty"`
}

// AdminClient is a Keycloak admin client.
type AdminClient struct {
	ServerURL string
	Realm     string
	Username  string
	Password  string
	ClientID  string
	UseTLS    bool

	client *http.Client
}

// NewAdminClient creates a new Keycloak admin client.
func NewAdminClient(ctx *context.Context, serverURL, username, password string) (*AdminClient, error) {
	adminClient := &AdminClient{
		ServerURL: serverURL,
		Realm:     masterRealm,
		Username:  username,
		Password:  password,
		ClientID:  adminClientID,
	}

	if (*ctx).Value(http.Client{}) == nil {
		adminClient.client = http.DefaultClient
	} else {
		adminClient.client = (*ctx).Value(http.Client{}).(*http.Client)
	}

	// test connection
	if _, err := adminClient.getToken(); err != nil {
		return nil, err
	}

	return adminClient, nil
}

// GetClient returns a Keycloak client.
func (a *AdminClient) GetClient(realm string, clientID string) (*Client, error) {
	token, err := a.getToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", a.ServerURL+"/admin/realms/"+realm+"/clients", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token.AccessToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var clients []Client
	if err = json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, err
	}

	for _, c := range clients {
		if *c.ClientID == clientID {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("client not found")
}

func (a *AdminClient) getToken() (*Token, error) {
	var token Token

	resp, err := a.client.PostForm(
		a.ServerURL+"/realms/"+a.Realm+"/protocol/openid-connect/token",
		url.Values{
			"grant_type": {"password"},
			"client_id":  {a.ClientID},
			"username":   {a.Username},
			"password":   {a.Password},
		},
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// ClientContext returns a new context with the given HTTP client
// Used to pass a custom HTTP client to the AdminClient
func ClientContext(ctx context.Context, client *http.Client) context.Context {
	return context.WithValue(ctx, http.Client{}, client)
}
