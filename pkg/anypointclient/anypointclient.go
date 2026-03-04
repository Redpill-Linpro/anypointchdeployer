package anypointclient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

const USRegionBaseURL = "https://anypoint.mulesoft.com"
const EURegionBaseURL = "https://eu1.anypoint.mulesoft.com"

type AuthenticationType string

const (
	BearerAuthenticationType       AuthenticationType = "bearer"
	UserAuthenticationType         AuthenticationType = "user"
	ConnectedAppAuthenticationType AuthenticationType = "connectedapp"
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
	authType     AuthenticationType
	baseURL      string
}

// createHTTPClient creates an HTTP client, optionally configured with a proxy
func createHTTPClient(proxyURL string) *http.Client {
	if proxyURL == "" {
		return &http.Client{}
	}

	proxy, err := url.Parse(proxyURL)
	if err != nil {
		// If proxy URL is invalid, fall back to default client
		return &http.Client{}
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}

	return &http.Client{
		Transport: transport,
	}
}

/*
NewAnypointClientWithToken creates a new Anypoint Client using the given token
*/
func NewAnypointClientWithToken(bearer string, baseURL string, proxyURL string) *AnypointClient {
	var c AnypointClient

	c.HTTPClient = createHTTPClient(proxyURL)
	c.bearer = bearer
	c.baseURL = baseURL
	c.authType = BearerAuthenticationType
	return &c
}

/*
NewAnypointClientWithCredentials creates a new Anypoint Client using the given username and password to acquire a token
*/
func NewAnypointClientWithCredentials(username string, password string, baseURL string, proxyURL string) *AnypointClient {
	var c AnypointClient

	c.HTTPClient = createHTTPClient(proxyURL)
	c.baseURL = baseURL
	c.username = username
	c.password = password
	c.bearer = ""
	c.authType = UserAuthenticationType

	return &c
}

/*
NewAnypointClientWithConnectedApp creates a new Anypoint Client using the given client id and client secret to acquire a token
*/
func NewAnypointClientWithConnectedApp(clientId string, clientSecret string, baseURL string, proxyURL string) *AnypointClient {
	var c AnypointClient

	c.HTTPClient = createHTTPClient(proxyURL)
	c.baseURL = baseURL
	c.clientId = clientId
	c.clientSecret = clientSecret
	c.bearer = ""
	c.authType = ConnectedAppAuthenticationType

	return &c
}

func (client *AnypointClient) newRequest(method string, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s/%s", client.baseURL, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if client.bearer != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.bearer))
	}
	return req, nil
}

func (client *AnypointClient) newGetRequest(path string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", client.baseURL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if client.bearer != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.bearer))
	}
	return client.HTTPClient.Do(req)
}

func ResolveBaseURLFromRegion(region string) (string, error) {
	switch strings.ToUpper(region) {
	case "EU":
		return EURegionBaseURL, nil
	case "US":
		return USRegionBaseURL, nil
	default:
		return "", errors.Errorf("%s is not a valid region", region)
	}
}
