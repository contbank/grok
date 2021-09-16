package grok

import (
	"net/http"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	//ErrClientIDClientSecret ...
	ErrClientIDClientSecret = NewError(http.StatusInternalServerError, "error client id or client secret")
)

//Config ...
type Config struct {
	LoginEndpoint *string
	APIEndpoint   *string
	ClientID      *string `validate:"required"`
	ClientSecret  *string `validate:"required"`
	Realm         *string
	GrantType     *string
	Username      *string `validate:"required"`
	Password      *string `validate:"required"`
	Audience      *string
	Scopes        *string
	APIVersion    *string
	Cache         *cache.Cache
}

//Session ...
type Session struct {
	LoginEndpoint string      `json:"login_endpoint,omitempty"`
	APIEndpoint   string      `json:"api_endpoint,omitempty"`
	ClientID      string      `json:"client_id,omitempty"`
	ClientSecret  string      `json:"client_secret,omitempty"`
	Realm         string      `json:"realm,omitempty"`
	GrantType     string      `json:"grant_type,omitempty"`
	Username      string      `json:"username,omitempty"`
	Password      string      `json:"password,omitempty"`
	Audience      string      `json:"audience,omitempty"`
	Scopes        string      `json:"scopes,omitempty"`
	APIVersion    string      `json:"apiversion,omitempty"`
	Cache         cache.Cache `json:"cache,omitempty"`
}

//NewSession ...
func NewSession(config Config) (*Session, error) {
	err := Validator.Struct(config)

	if err != nil {
		return nil, FromValidationErros(err)
	}

	if config.LoginEndpoint == nil {
		config.LoginEndpoint = String("https://contbank.us.auth0.com/oauth/token")
	}
	if config.APIEndpoint == nil {
		config.APIEndpoint = String("")
	}

	if config.APIVersion == nil {
		config.APIVersion = String("1.0")
	}
	if config.ClientID == nil {
		config.ClientID = String(os.Getenv("INTRA_CLIENT_ID"))
	}
	if config.ClientSecret == nil {
		config.ClientID = String(os.Getenv("INTRA_CLIENT_SECRET"))
	}
	if config.GrantType == nil {
		config.GrantType = String("http://auth0.com/oauth/grant-type/password-realm")
	}
	if config.Realm == nil {
		config.Realm = String("Username-Password-Authentication")
	}
	if config.Username == nil {
		return nil, ErrClientIDClientSecret
	}
	if config.Password == nil {
		return nil, ErrClientIDClientSecret
	}
	if config.Audience == nil {
		config.Audience = String("https://api.contbank.com")
	}
	if config.Scopes == nil {
		config.Scopes = String("openid email offline_access")
	}
	if *config.ClientID == "" || *config.ClientSecret == "" {
		return nil, ErrClientIDClientSecret
	}

	if config.Cache == nil {
		config.Cache = cache.New(10*time.Minute, 1*time.Second)
	}

	var session = &Session{
		GrantType:     *config.GrantType,
		Realm:         *config.Realm,
		Username:      *config.Username,
		Password:      *config.Password,
		Audience:      *config.Audience,
		Scopes:        *config.Scopes,
		APIEndpoint:   *config.APIEndpoint,
		ClientID:      *config.ClientID,
		ClientSecret:  *config.ClientSecret,
		APIVersion:    *config.APIVersion,
		LoginEndpoint: *config.LoginEndpoint,
		Cache:         *config.Cache,
	}

	return session, nil
}
