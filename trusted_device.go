package grok

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	DeviceIDHeader        = "X-Device-Id"
	DeviceNonceHeader     = "X-Device-Nonce"
	DeviceSignatureHeader = "X-Device-Signature"
)

type TrustedDevice interface {
	Validate() gin.HandlerFunc
}

type TrustedDeviceSettings struct {
	Fake    bool   `yaml:"fake"`
	URL     string `yaml:"url"`
	Success *bool  `yaml:"success"`
}

type InternalTrustedDevice struct {
	settings *TrustedDeviceSettings
}

type trustedDeviceVerifyPayload struct {
	UserID      string `json:"user_id"`
	DeviceID    string `json:"device_id"`
	Nonce       string `json:"nonce"`
	Signature   string `json:"signature"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	PayloadHash string `json:"payload_hash"`
}

func CreateTrustedDevice(settings *TrustedDeviceSettings) TrustedDevice {
	if settings == nil || settings.Fake {
		success := true
		if settings != nil && settings.Success != nil {
			success = *settings.Success
		}
		return NewFakeTrustedDevice(success)
	}

	return NewInternalTrustedDevice(settings)
}

func NewInternalTrustedDevice(settings *TrustedDeviceSettings) TrustedDevice {
	return &InternalTrustedDevice{settings: settings}
}

func (a *InternalTrustedDevice) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		defaultError := Error{
			Code:     http.StatusForbidden,
			Key:      "DEVICE_NOT_TRUSTED",
			Messages: []string{"aparelho não cadastrado ou sem permissão"},
		}

		if a.settings == nil || strings.TrimSpace(a.settings.URL) == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		userID := strings.TrimSpace(c.GetString("user_id"))
		deviceID := strings.TrimSpace(c.GetHeader(DeviceIDHeader))
		nonce := strings.TrimSpace(c.GetHeader(DeviceNonceHeader))
		signature := strings.TrimSpace(c.GetHeader(DeviceSignatureHeader))
		if userID == "" || deviceID == "" || nonce == "" || signature == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		sum := sha256.Sum256(bodyBytes)
		payload := trustedDeviceVerifyPayload{
			UserID:      userID,
			DeviceID:    deviceID,
			Nonce:       nonce,
			Signature:   signature,
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			PayloadHash: hex.EncodeToString(sum[:]),
		}

		b, err := json.Marshal(payload)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		req, err := http.NewRequest("POST", a.settings.URL, bytes.NewReader(b))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", c.Request.Header.Get("authorization"))
		if identity := c.GetHeader(CurrentIdentityHeader); identity != "" {
			req.Header.Set(CurrentIdentityHeader, identity)
		}

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			c.Next()
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil || len(body) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		var response Error
		if err := json.Unmarshal(body, &response); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		status := resp.StatusCode
		if response.Code > 0 {
			status = response.Code
		}
		c.AbortWithStatusJSON(status, response)
	}
}
