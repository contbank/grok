package grok

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

//IntraAuthentication ...
type IntraAuthentication struct {
	session    Session
	httpClient *http.Client
}

// IntraAuthenticationRequest ...
type IntraAuthenticationRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Realm        string `json:"realm"`
	GrantType    string `json:"grant_type"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Audience     string `json:"audience"`
	Scopes       string `json:"scopes"`
}

//NewIntraAuthentication ...
func NewIntraAuthentication(session Session) *IntraAuthentication {
	return &IntraAuthentication{
		session: session,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

var (
	// ErrDefaultLogin ...
	ErrDefaultLogin = NewError(http.StatusInternalServerError, "error login")
)

// AuthenticationResponse ...
type IntraAuthenticationResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// ErrorLoginResponse ...
type ErrorLoginResponse struct {
	Message string `json:"error"`
}

func (a *IntraAuthentication) login(ctx context.Context, model IntraAuthenticationRequest) (*IntraAuthenticationResponse, error) {
	u, err := url.Parse(a.session.LoginEndpoint)

	if err != nil {
		return nil, err
	}

	//u.Path = path.Join(u.Path, LoginPath)
	endpoint := u.String()

	//reqbyte for body json format
	reqbyte, err := json.Marshal(a.session)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(reqbyte))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-type", "application/json")

	resp, err := a.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var response *IntraAuthenticationResponse

		respBody, _ := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(respBody, &response)

		if err != nil {
			return nil, err
		}
		return response, nil
	}
	if resp.StatusCode == http.StatusBadRequest {
		var bodyErr *ErrorLoginResponse

		respBody, _ := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(respBody, &bodyErr)

		if err != nil {
			return nil, err

		}
		logrus.
			WithError(err).
			Errorf("400")
		return nil, FromValidationErros(err)
		//return nil, FindError("400", bodyErr.Message)
	}

	return nil, ErrDefaultLogin
}

//Token ...
func (a IntraAuthentication) Token(ctx context.Context, model IntraAuthenticationRequest) (string, error) {
	if token, found := a.session.Cache.Get("intra_access_token"); found {
		return token.(string), nil
	}

	response, err := a.login(ctx, model)

	if err != nil {
		return "", err
	}

	a.session.Cache.Set("intra_access_token", fmt.Sprintf("%s %s", response.TokenType, response.AccessToken), time.Second*time.Duration(int64(response.ExpiresIn-10)))
	ctx = context.WithValue(ctx, "intra_access_token", response.AccessToken)

	return fmt.Sprintf("%s %s", response.TokenType, response.AccessToken), nil
}
