package anypointclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type Organization struct {
	ID               string
	Name             string
	SubOrganizations []Organization
}

var organizationCache struct {
	value  Organization
	loaded bool
}

func (client *AnypointClient) ResolveOrganization(organizationPath string) (Organization, error) {
	org, err := client.getOrganizationTree()
	if err != nil {
		return Organization{}, errors.Wrapf(err, "failed to find organtization %s", organizationPath)
	}
	parts := strings.Split(organizationPath, "/")
	if parts[0] == org.Name {
		for _, part := range parts[1:] {
			org = *org.findChild(part)
		}
	} else {
		return Organization{}, fmt.Errorf("failed to find organtization %s", organizationPath)
	}
	return org, err
}

func (organization *Organization) findChild(name string) *Organization {
	for _, subOrg := range organization.SubOrganizations {
		if subOrg.Name == name {
			return &subOrg
		}
	}
	return nil
}

func (client *AnypointClient) getOrganizationTree() (Organization, error) {
	if !organizationCache.loaded {
		req, _ := client.newRequest("GET", "accounts/api/me", nil)
		res, err := client.HTTPClient.Do(req)
		if err != nil {
			return Organization{}, err
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusOK {
			err := organizationCache.value.buildOrganizationTree(res.Body)
			if err != nil {
				return Organization{}, err
			}
			organizationCache.loaded = true
		}
	}
	return organizationCache.value, nil
}

func (organization *Organization) buildOrganizationTree(body io.ReadCloser) error {
	var userInfoJSON map[string]any

	dec := json.NewDecoder(body)
	if err := dec.Decode(&userInfoJSON); err != nil {
		return err
	}

	// TODO: Add checks and bounds
	organization.ID = userInfoJSON["user"].(map[string]any)["organization"].(map[string]any)["id"].(string)
	organization.Name = userInfoJSON["user"].(map[string]any)["organization"].(map[string]any)["name"].(string)

	// TODO: Add checks and bounds
	organizationsJSON := userInfoJSON["user"].(map[string]any)["memberOfOrganizations"].([]any)
	organization.buildRecursiveFromJSON(organizationsJSON)
	return nil
}

func (organization *Organization) buildRecursiveFromJSON(organizationsJSON []any) error {
	// TODO: Add checks and bounds
	for _, val := range organizationsJSON {
		// Check for the organization we are looking for
		if val.(map[string]any)["id"].(string) == organization.ID {
			organization.Name = val.(map[string]any)["name"].(string)
			// Check if it has subOrganizations
			for _, val2 := range val.(map[string]any)["subOrganizationIds"].([]any) {
				subOrg := Organization{
					ID: val2.(string),
				}
				subOrg.buildRecursiveFromJSON(organizationsJSON)
				organization.SubOrganizations = append(organization.SubOrganizations, subOrg)
			}
		}
	}
	return nil
}
