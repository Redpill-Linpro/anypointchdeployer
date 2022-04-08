package anypointclient

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

/*
AnypointClient represents the
*/
type AnypointClient struct {
	HTTPClient *http.Client
	Username   string
	secret     string
	bearer     string
	baseURL    string
}

/*
NewAnypointClientWithToken creates a new Anypoint Client using the given token
*/
func NewAnypointClientWithToken(baseURL string, bearer string) *AnypointClient {
	var c AnypointClient

	c.HTTPClient = &http.Client{}
	c.bearer = bearer
	c.baseURL = resolveBaseURLFromRegion(baseURL)
	return &c
}

/*
NewAnypointClientWithCredentials creates a new Anypoint Client using the given username and password to aquire a token
*/
func NewAnypointClientWithCredentials(baseURL string, username string, secret string) *AnypointClient {
	var c AnypointClient

	c.HTTPClient = &http.Client{}
	c.baseURL = resolveBaseURLFromRegion(baseURL)
	c.Username = username
	c.secret = secret
	c.bearer = c.getAuthorizationBearerToken()

	return &c
}

func (client *AnypointClient) newRequest(method string, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s/%s",
		client.baseURL,
		path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if client.bearer != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.bearer))
	}
	return req, nil
}

func resolveBaseURLFromRegion(region string) string {
	switch strings.ToUpper(region) {
	case "EU":
		return "https://eu1.anypoint.mulesoft.com"
	case "US":
		return "https://anypoint.mulesoft.com"
	default:
		return region
	}
}
