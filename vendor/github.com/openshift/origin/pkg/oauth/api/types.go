package api

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
)

type AccessToken struct {
	api.JSONBase `json:",inline" yaml:",inline"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// Name is the unique value for an access token - also known as its secret
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// AuthorizeToken is the authorization token that granted this access token, and contains
	// the specific state of the token.
	AuthorizeToken AuthorizeToken `json:"authorizeToken,omitempty" yaml:"authorizeToken,omitempty"`

	// RefreshToken is the value by which this token can be renewed. Can be blank.
	RefreshToken string `json:"refreshToken,omitempty" yaml:"refreshToken,omitempty"`
}

type AuthorizeToken struct {
	api.JSONBase `json:",inline" yaml:",inline"`

	// Name is the unique value for an authorization token - also known as its secret
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// ClientName references the client that created this token.
	ClientName string `json:"clientName,omitempty" yaml:"clientName,omitempty"`

	// ExpiresIn is the seconds from CreationTime before this token expires.
	ExpiresIn int64 `json:"expiresIn,omitempty" yaml:"expiresIn,omitempty"`

	// Scopes is an array of the requested scopes.
	Scopes []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`

	// RedirectURI is the redirection associated with the token.
	RedirectURI string `json:"redirectURI,omitempty" yaml:"redirectURI,omitempty"`

	// State data from request
	State string `json:"state,omitempty" yaml:"state,omitempty"`

	// UserName is the user name associated with this token
	UserName string `json:"userName,omitempty" yaml:"userName,omitempty"`

	// UserUID is the unique UID associated with this token. UserUID and UserName must both match
	// for this token to be valid.
	UserUID string `json:"userUID,omitempty" yaml:"userUID,omitempty"`
}

type Client struct {
	api.JSONBase `json:",inline" yaml:",inline"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// Name is the unique identifier of the client
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Secret is the unique secret associated with a client
	Secret string `json:"secret,omitempty" yaml:"secret,omitempty"`

	// RedirectURIs is the valid redirection URIs associated with a client
	RedirectURIs []string `json:"redirectURIs,omitempty" yaml:"redirectURIs,omitempty"`
}

type ClientAuthorization struct {
	api.JSONBase `json:",inline" yaml:",inline"`

	// ClientName references the client that created this authorization
	ClientName string `json:"clientName,omitempty" yaml:"clientName,omitempty"`

	// UserName is the user name that authorized this client
	UserName string `json:"userName,omitempty" yaml:"userName,omitempty"`

	// UserUID is the unique UID associated with this authorization. UserUID and UserName
	// must both match for this authorization to be valid.
	UserUID string `json:"userUID,omitempty" yaml:"userUID,omitempty"`

	// Scopes is an array of the granted scopes.
	Scopes []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

type AccessTokenList struct {
	api.JSONBase `json:",inline" yaml:",inline"`
	Items        []AccessToken `json:"items,omitempty" yaml:"items,omitempty"`
}

type AuthorizeTokenList struct {
	api.JSONBase `json:",inline" yaml:",inline"`
	Items        []AuthorizeToken `json:"items,omitempty" yaml:"items,omitempty"`
}

type ClientList struct {
	api.JSONBase `json:",inline" yaml:",inline"`
	Items        []Client `json:"items,omitempty" yaml:"items,omitempty"`
}

type ClientAuthorizationList struct {
	api.JSONBase `json:",inline" yaml:",inline"`
	Items        []ClientAuthorization `json:"items,omitempty" yaml:"items,omitempty"`
}

func (*AccessToken) IsAnAPIObject()             {}
func (*AuthorizeToken) IsAnAPIObject()          {}
func (*Client) IsAnAPIObject()                  {}
func (*AccessTokenList) IsAnAPIObject()         {}
func (*AuthorizeTokenList) IsAnAPIObject()      {}
func (*ClientList) IsAnAPIObject()              {}
func (*ClientAuthorization) IsAnAPIObject()     {}
func (*ClientAuthorizationList) IsAnAPIObject() {}
