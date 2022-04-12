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
	HTTPClient   *http.Client
	username     string
	password     string
	clientId     string
	clientSecret string
	bearer       string
	baseURL      string
}

const USRegionBaseURL = "https://anypoint.mulesoft.com"
const EURegionBaseURL = "https://eu1.anypoint.mulesoft.com"

/*
NewAnypointClientWithToken creates a new Anypoint Client using the given token
*/
func NewAnypointClientWithToken(region string, bearer string) *AnypointClient {
	var c AnypointClient

	c.HTTPClient = &http.Client{}
	c.bearer = bearer
	c.baseURL = resolveBaseURLFromRegion(region)
	return &c
}

/*
NewAnypointClientWithCredentials creates a new Anypoint Client using the given username and password to acquire a token
*/
func NewAnypointClientWithCredentials(region string, username string, password string) *AnypointClient {
	var c AnypointClient

	c.HTTPClient = &http.Client{}
	c.baseURL = resolveBaseURLFromRegion(region)
	c.username = username
	c.password = password
	c.bearer = c.getAuthorizationBearerToken("user")

	return &c
}

/*
NewAnypointClientWithCredentials creates a new Anypoint Client using the given client id and client secret to acquire a token
*/
func NewAnypointClientWithConnectedApp(region string, clientId string, clientSecret string) *AnypointClient {
	var c AnypointClient

	c.HTTPClient = &http.Client{}
	c.baseURL = resolveBaseURLFromRegion(region)
	c.clientId = clientId
	c.clientSecret = clientSecret
	c.bearer = c.getAuthorizationBearerToken("connectedapp")

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
		return EURegionBaseURL
	case "US":
		return USRegionBaseURL
	default:
		return region
	}
}
