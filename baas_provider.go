package grok

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/gin-gonic/gin"
)

const (
	// X_BAAS_PROVIDER ...
	X_BAAS_PROVIDER string = "X-Baas-Provider"
	// DEFAULT_PROVIDER ...
	DEFAULT_PROVIDER string = "CELCOIN"
	// BANKLY_PROVIDER ...
	BANKLY_PROVIDER string = "BANKLY"
	// CELCOIN_PROVIDER ...
	CELCOIN_PROVIDER string = "CELCOIN"
)

// BaasProvider ...
type BaasProvider interface {
	Identify() gin.HandlerFunc
}

// BaasProviderIntra ...
type BaasProviderIntra interface {
	Identify() gin.HandlerFunc
}

// baasProvider ...
type baasProvider struct {
	settings *BaasProviderSettings
}

// baasProviderIntra ...
type baasProviderIntra struct {
	settings *BaasProviderIntraSettings
}

// NewBaasProvider ...
func NewBaasProvider(settings *BaasProviderSettings) BaasProvider {
	return &baasProvider{
		settings: settings,
	}
}

// CreateBaasProvider ...
func CreateBaasProvider(settings *BaasProviderSettings) BaasProvider {
	if settings.Fake {
		success := true
		if settings.Success != nil {
			success = *settings.Success
		}
		return NewFakeBaasProvider(success)
	}
	return NewBaasProvider(settings)
}

// NewBaasProviderIntra ...
func NewBaasProviderIntra(settings *BaasProviderIntraSettings) BaasProviderIntra {
	return &baasProviderIntra{
		settings: settings,
	}
}

// CreateBaasProviderIntra ...
func CreateBaasProviderIntra(settings *BaasProviderIntraSettings) BaasProviderIntra {
	if settings.Fake {
		success := true
		if settings.Success != nil {
			success = *settings.Success
		}
		return NewFakeBaasProvider(success)
	}
	return NewBaasProviderIntra(settings)
}

// Identify ...
func (p *baasProvider) Identify() gin.HandlerFunc {
	return func(c *gin.Context) {
		if p.settings == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// current identity is required
		currentIdentity := c.Request.Header.Get(X_CURRENT_IDENTITY)
		if len(currentIdentity) == 0 {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		url := p.settings.URL
		jwt := c.Request.Header.Get("authorization")

		baasProvider := p.getBaasProviderForIdentifier(&currentIdentity, url, &jwt)

		c.Set(X_BAAS_PROVIDER, *baasProvider)

		c.Next()
	}
}

// Identify ...
func (p *baasProviderIntra) Identify() gin.HandlerFunc {
	return func(c *gin.Context) {
		if p.settings == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// current identity is required
		currentIdentity := c.Request.Header.Get(X_CURRENT_IDENTITY)
		if len(currentIdentity) == 0 {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		url := p.settings.URL
		jwt := c.Request.Header.Get("authorization")

		baasProvider := p.getBaasProviderIntraForIdentifier(&currentIdentity, url, &jwt)

		c.Set(X_BAAS_PROVIDER, *baasProvider)

		c.Next()
	}
}

// getBaasProviderForIdentifier ...
func (p *baasProvider) getBaasProviderForIdentifier(identifier *string, endpoint *string, jwt *string) *string {

	baasProvider := DEFAULT_PROVIDER

	u, err := url.Parse(*endpoint)
	if err != nil {
		return &baasProvider
	}

	u.Path = path.Join(u.Path, *identifier)
	newEndpoint := u.String()

	req, err := http.NewRequest("GET", newEndpoint, nil)
	if err != nil {
		return &baasProvider
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", *jwt)
	req.Header.Set(X_CURRENT_IDENTITY, *identifier)

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return &baasProvider
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &baasProvider
	}

	response := struct {
		BaasProvider string `json:"baas_provider,omitempty"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return &baasProvider
	}
	return &response.BaasProvider
}

// getBaasProviderIntraForIdentifier ...
func (p *baasProviderIntra) getBaasProviderIntraForIdentifier(identifier *string, endpoint *string, jwt *string) *string {

	baasProvider := DEFAULT_PROVIDER

	u, err := url.Parse(*endpoint)
	if err != nil {
		return &baasProvider
	}

	u.Path = path.Join(u.Path, *identifier)
	newEndpoint := u.String()

	req, err := http.NewRequest("GET", newEndpoint, nil)
	if err != nil {
		return &baasProvider
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", *jwt)
	req.Header.Set(X_CURRENT_IDENTITY, *identifier)

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return &baasProvider
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &baasProvider
	}

	response := struct {
		BaasProvider string `json:"baas_provider,omitempty"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return &baasProvider
	}
	return &response.BaasProvider
}

// FakeBaasProvider ...
type FakeBaasProvider struct {
	alwaysSuccess bool
}

// NewFakeAuthorize ...
func NewFakeBaasProvider(success bool) BaasProvider {
	return &FakeBaasProvider{
		alwaysSuccess: success,
	}
}

// Identify ...
func (a *FakeBaasProvider) Identify() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.alwaysSuccess {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Set(X_BAAS_PROVIDER, DEFAULT_PROVIDER)
		c.Next()
	}
}

// IsBanklyProvider ...
func IsBanklyProvider(baasProvider *string) bool {
	return baasProvider != nil && *baasProvider == BANKLY_PROVIDER
}

// IsCelcoinProvider ...
func IsCelcoinProvider(baasProvider *string) bool {
	return baasProvider != nil && *baasProvider == CELCOIN_PROVIDER
}
