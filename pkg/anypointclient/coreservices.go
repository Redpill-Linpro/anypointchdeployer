package anypointclient

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

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

func (client *AnypointClient) getAuthorizationBearerToken() (token string) {
	loginURL :=
		"accounts/login"

	data := url.Values{}
	data.Set("username", client.Username)
	data.Set("password", client.password)

	req, _ := client.newRequest("POST", loginURL, strings.NewReader(data.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

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
