package anypointclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Environment struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organizationId"`
	IsProduction   bool   `json:"isProduction"`
	EnvType        string `json:"type"`
	ClientID       string `json:"clientId"`
}

type EnvironmentResponse struct {
	Data  []Environment `json:"data"`
	Total int           `json:"total"`
}

/*
ResolveEnvironment will resolve, in the given organisation, an Environment by name.
*/
func (client *AnypointClient) ResolveEnvironment(organization Organization, environmentName string) (Environment, error) {
	envResp := new(EnvironmentResponse)

	req, _ := client.newRequest("GET", fmt.Sprintf("/accounts/api/organizations/%s/environments", organization.ID), nil)
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return Environment{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return Environment{}, err
		}
		err = json.Unmarshal(bodyBytes, &envResp)
		if err != nil {
			return Environment{}, err
		}
	}

	for _, e := range envResp.Data {
		if e.Name == environmentName {
			return e, nil
		}
	}
	return Environment{}, fmt.Errorf("failed to find environment named %s in organisation %s", environmentName, organization.Name)
}
