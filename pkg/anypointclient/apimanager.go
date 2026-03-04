package anypointclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type ApiListResponse struct {
	Total     int `json:"total"`
	Instances []struct {
		Audit struct {
			Created struct {
				Date time.Time `json:"date"`
			} `json:"created"`
			Updated struct {
				Date time.Time `json:"date"`
			} `json:"updated"`
		} `json:"audit"`
		MasterOrganizationID string    `json:"masterOrganizationId"`
		OrganizationID       string    `json:"organizationId"`
		ID                   int       `json:"id"`
		InstanceLabel        string    `json:"instanceLabel"`
		GroupID              string    `json:"groupId"`
		AssetID              string    `json:"assetId"`
		AssetVersion         string    `json:"assetVersion"`
		ProductVersion       string    `json:"productVersion"`
		Description          any       `json:"description"`
		Tags                 []any     `json:"tags"`
		Order                int       `json:"order"`
		ProviderID           any       `json:"providerId"`
		Deprecated           bool      `json:"deprecated"`
		LastActiveDate       time.Time `json:"lastActiveDate"`
		EndpointURI          string    `json:"endpointUri"`
		EnvironmentID        string    `json:"environmentId"`
		IsPublic             bool      `json:"isPublic"`
		Stage                string    `json:"stage"`
		Technology           string    `json:"technology"`
		LastActiveDelta      int       `json:"lastActiveDelta,omitempty"`
		Pinned               bool      `json:"pinned"`
		ActiveContractsCount int       `json:"activeContractsCount"`
		Asset                struct {
			Name              string `json:"name"`
			ExchangeAssetName string `json:"exchangeAssetName"`
			GroupID           string `json:"groupId"`
			AssetID           string `json:"assetId"`
		} `json:"asset"`
		AutodiscoveryInstanceName string `json:"autodiscoveryInstanceName"`
	} `json:"instances"`
}

type ApiPolicyResponse struct {
	Audit struct {
		Created struct {
			Date time.Time `json:"date"`
		} `json:"created"`
		Updated struct {
			Date time.Time `json:"date"`
		} `json:"updated"`
	} `json:"audit"`
	MasterOrganizationID string         `json:"masterOrganizationId,omitempty"`
	OrganizationID       string         `json:"organizationId,omitempty"`
	PolicyTemplateID     string         `json:"policyTemplateId,omitempty"`
	Configuration        map[string]any `json:"configuration,omitempty"`
	Order                int            `json:"order,omitempty"`
	Disabled             bool           `json:"disabled,omitempty"`
	PointcutData         any            `json:"pointcutData,omitempty"`
	Template             struct {
		GroupID      string `json:"groupId,omitempty"`
		AssetID      string `json:"assetId,omitempty"`
		AssetVersion string `json:"assetVersion,omitempty"`
	} `json:"template"`
	Standalone          bool                `json:"standalone,omitempty"`
	APIID               int                 `json:"apiId,omitempty"`
	ImplementationAsset ImplementationAsset `json:"implementationAsset"`
	Type                string              `json:"type,omitempty"`
	PolicyID            int                 `json:"policyId,omitempty"`
	Version             int64               `json:"version,omitempty"`
}

type ApiPolicyRequest struct {
	ConfigurationData map[string]any `json:"configurationData,omitempty"`
	Order             int            `json:"order,omitempty"`
	Disabled          bool           `json:"disabled,omitempty"`
	PointcutData      any            `json:"pointcutData,omitempty"`
	GroupID           string         `json:"groupId,omitempty"`
	AssetID           string         `json:"assetId,omitempty"`
	AssetVersion      string         `json:"assetVersion,omitempty"`
	Standalone        bool           `json:"standalone,omitempty"`
}

type ImplementationAsset struct {
	GroupID               string   `json:"groupId"`
	AssetID               string   `json:"assetId"`
	Version               string   `json:"version"`
	Technology            string   `json:"technology"`
	MinRuntimeVersion     string   `json:"minRuntimeVersion"`
	SupportedJavaVersions []string `json:"supportedJavaVersions,omitempty"`
}

func (client *AnypointClient) GetApis(orgId string, envId string, offset int, limit int) (*ApiListResponse, error) {

	getAPIUrl := fmt.Sprintf(
		"apimanager/xapi/v1/organizations/%s/environments/%s/apis?ascending=true&limit=%d&offset=%d&sort=name",
		orgId,
		envId,
		limit,
		offset,
	)
	req, _ := client.newRequest("GET", getAPIUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.bearer))
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var response map[string]any
		bodyBytes, err := io.ReadAll(res.Body)
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return nil,
				errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)

		}
		return nil, errors.Errorf("Call to Anypoint Platform returned %d : %s",
			res.StatusCode, response["message"])
	}

	var response ApiListResponse
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read response from Anypoint Platform")
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal response from Anypoint Platform")
	}
	return &response, nil
}

func (client *AnypointClient) GetApiInstancePolicies(orgId string, envId string, apiInstanceID int) (*[]ApiPolicyResponse, error) {
	getAPIInstancePolicyURL := fmt.Sprintf(
		"apimanager/api/v1/organizations/%s/environments/%s/apis/%d/policies?fullInfo=true",
		orgId,
		envId,
		apiInstanceID,
	)
	req, _ := client.newRequest("GET", getAPIInstancePolicyURL, nil)
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	var response struct {
		Policies []ApiPolicyResponse
	} = struct{ Policies []ApiPolicyResponse }{}
	if res.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read response from Anypoint Platform")
		}
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			fmt.Printf("Failed to unmarshal response from Anypoint Platform: %v %s\n", err, string(bodyBytes))
			return nil, errors.Wrapf(err, "failed to unmarshal response from Anypoint Platform")
		}
	}
	return &response.Policies, nil
}

func (client *AnypointClient) UpdateApiInstancePolicies(orgId string, envId string, apiInstanceID int, policyID int, apipolicy ApiPolicyRequest) error {
	updateAPIInstancePolicyURL := fmt.Sprintf(
		"apimanager/api/v1/organizations/%s/environments/%s/apis/%d/policies/%d",
		orgId,
		envId,
		apiInstanceID,
		policyID,
	)
	updatePolicyPayload, err := json.Marshal(apipolicy)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal API Policy to JSON")
	}
	req, _ := client.newRequest("PATCH", updateAPIInstancePolicyURL, bytes.NewBuffer([]byte(updatePolicyPayload)))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "failed to read response from Anypoint Platform")
		}
		return errors.Errorf("Failed to update API Instance Policies status code: %d\nResponse Message:%s",
			res.StatusCode, string(bodyBytes))
	}

	return nil
}

func (client *AnypointClient) CreateApiInstancePolicies(orgId string, envId string, apiInstanceID int, apipolicy ApiPolicyRequest) error {
	createAPIInstancePolicyURL := fmt.Sprintf(
		"apimanager/api/v1/organizations/%s/environments/%s/apis/%d/policies",
		orgId,
		envId,
		apiInstanceID,
	)
	updatePolicyPayload, err := json.Marshal(apipolicy)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal API Policy to JSON")
	}
	req, _ := client.newRequest("POST", createAPIInstancePolicyURL, bytes.NewBuffer([]byte(updatePolicyPayload)))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "failed to read response from Anypoint Platform")
		}
		return errors.Errorf("Failed to update API Instance Policies status code: %d\nResponse Message:%s",
			res.StatusCode, string(bodyBytes))
	}

	return nil
}
