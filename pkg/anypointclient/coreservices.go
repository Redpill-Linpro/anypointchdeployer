package anypointclient

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

const connectedAppLoginURL = "accounts/api/v2/oauth2/token"
const userLoginURL = "accounts/login"

/*
LoginRequest represents the form data beeing send in the Login request.
*/
type LoginRequest struct {
	Username string `url:"username"`
	Password string `url:"password"`
}

/*
LoginResponse represents the JSON data beeing returned by the Login request.
*/
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	RedirectURL string `url:"redirectUrl"`
}

func (client *AnypointClient) Login() error {
	var err error
	// We are already "logged in"
	if client.authType == BearerAuthenticationType {
		return nil
	}
	client.bearer, err = client.getAuthorizationBearerToken(client.authType)
	return err
}

func (client *AnypointClient) getAuthorizationBearerToken(authType AuthenticationType) (string, error) {
	var loginURL string
	data := url.Values{}
	switch authType {
	case UserAuthenticationType:
		loginURL = userLoginURL

		data.Set("username", client.username)
		data.Set("password", client.password)
		break
	case ConnectedAppAuthenticationType:
		loginURL = connectedAppLoginURL

		data.Set("client_id", client.clientId)
		data.Set("client_secret", client.clientSecret)
		data.Set("grant_type", "client_credentials")
	default:
		return "", errors.Errorf("can not get bearer token using authentication type %s", authType)
	}

	req, _ := client.newRequest("POST", loginURL, strings.NewReader(data.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	var loginRespone LoginResponse

	if res.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", errors.Wrap(err, "failed to read response from Anypoint Platform")
		}
		err = json.Unmarshal(bodyBytes, &loginRespone)
		if err != nil {
			return "", errors.Wrap(err, "failed to unmarshal response from Anypoint Platform")
		}
	}

	return loginRespone.AccessToken, nil
}
