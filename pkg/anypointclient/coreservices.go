package anypointclient

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
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

func (client *AnypointClient) getAuthorizationBearerToken(authType string) (token string) {
	var loginURL string
	data := url.Values{}
	if authType == "user" {
		loginURL = userLoginURL

		data.Set("username", client.username)
		data.Set("password", client.password)

	} else {
		loginURL = connectedAppLoginURL

		data.Set("client_id", client.clientId)
		data.Set("client_secret", client.clientSecret)
		data.Set("grant_type", "client_credentials")
	}

	req, _ := client.newRequest("POST", loginURL, strings.NewReader(data.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var loginRespone LoginResponse

	if res.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(bodyBytes, &loginRespone)
		if err != nil {
			log.Fatal(err)
		}
	}

	return loginRespone.AccessToken
}
