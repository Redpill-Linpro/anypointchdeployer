package anypointclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"
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
		MasterOrganizationID string        `json:"masterOrganizationId"`
		OrganizationID       string        `json:"organizationId"`
		ID                   int           `json:"id"`
		InstanceLabel        string        `json:"instanceLabel"`
		GroupID              string        `json:"groupId"`
		AssetID              string        `json:"assetId"`
		AssetVersion         string        `json:"assetVersion"`
		ProductVersion       string        `json:"productVersion"`
		Description          interface{}   `json:"description"`
		Tags                 []interface{} `json:"tags"`
		Order                int           `json:"order"`
		ProviderID           interface{}   `json:"providerId"`
		Deprecated           bool          `json:"deprecated"`
		LastActiveDate       time.Time     `json:"lastActiveDate"`
		EndpointURI          string        `json:"endpointUri"`
		EnvironmentID        string        `json:"environmentId"`
		IsPublic             bool          `json:"isPublic"`
		Stage                string        `json:"stage"`
		Technology           string        `json:"technology"`
		LastActiveDelta      int           `json:"lastActiveDelta,omitempty"`
		Pinned               bool          `json:"pinned"`
		ActiveContractsCount int           `json:"activeContractsCount"`
		Asset                struct {
			Name              string `json:"name"`
			ExchangeAssetName string `json:"exchangeAssetName"`
			GroupID           string `json:"groupId"`
			AssetID           string `json:"assetId"`
		} `json:"asset"`
		AutodiscoveryInstanceName string `json:"autodiscoveryInstanceName"`
	} `json:"instances"`
}

type ApiPolicy struct {
	Audit struct {
		Created struct {
			Date time.Time `json:"date"`
		} `json:"created"`
		Updated struct {
			Date time.Time `json:"date"`
		} `json:"updated"`
	} `json:"audit"`
	MasterOrganizationID string            `json:"masterOrganizationId"`
	OrganizationID       string            `json:"organizationId"`
	ID                   int               `json:"id"`
	PolicyTemplateID     string            `json:"policyTemplateId"`
	ConfigurationData    ConfigurationData `json:"configurationData,omitempty"`
	Order                int               `json:"order"`
	Disabled             bool              `json:"disabled"`
	PointcutData         interface{}       `json:"pointcutData"`
	GroupID              interface{}       `json:"groupId"`
	AssetID              interface{}       `json:"assetId"`
	AssetVersion         interface{}       `json:"assetVersion"`
	Template             string            `json:"template"`
	Standalone           bool              `json:"standalone"`
	APIID                int               `json:"apiId"`
}

type ConfigurationData struct {
	PolicyID                   int                    `json:"policyId"`
	EndpointURI                string                 `json:"endpointUri"`
	IsRamlEndpoint             bool                   `json:"isRamlEndpoint"`
	IsWsdlEndpoint             bool                   `json:"isWsdlEndpoint"`
	IsHTTPEndpoint             bool                   `json:"isHttpEndpoint"`
	APIID                      int                    `json:"apiId"`
	Order                      int                    `json:"order"`
	AutodiscoveryAPIName       string                 `json:"autodiscoveryApiName"`
	AutodiscoveryInstanceName  string                 `json:"autodiscoveryInstanceName"`
	AdditionalPolicyProperties map[string]interface{} `json:"-"` // Rest of the fields should go here.
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
		return nil, errors.Wrapf(err, "Failed to call Anypoint Platform: %w")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var response map[string]interface{}
		bodyBytes, err := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return nil,
				errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)

		}
		return nil, Errorf("Call to Anypoint Platform returned %d : %s",
			res.StatusCode, response["message"])
	}

	var response ApiListResponse
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		log.Fatal(err)
	}
	return &response, nil
}

func (client *AnypointClient) GetApiInstancePolicies(orgId string, envId string, apiInstanceID int) *[]ApiPolicy {
	getAPIInstancePolicyURL := fmt.Sprintf(
		"apimanager/api/v1/organizations/%s/environments/%s/apis/%d/policies?fullInfo=true",
		orgId,
		envId,
		apiInstanceID,
	)
	req, _ := client.newRequest("GET", getAPIInstancePolicyURL, nil)
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var response []ApiPolicy
	if res.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			log.Fatal(err)
		}
	}
	return &response
}

func (client *AnypointClient) UpdateApiInstancePolicies(orgId string, envId string, apiInstanceID int,
	configData string, policyID int, newPolicyTemplateID int) error {
	updateAPIInstancePolicyURL := fmt.Sprintf(
		"apimanager/api/v1/organizations/%s/environments/%s/apis/%d/policies/%d",
		orgId,
		envId,
		apiInstanceID,
		policyID,
	)

	updatePolicyPayload :=
		fmt.Sprintf(
			`{"configurationData": %s,"pointcutData":null,"policyTemplateId":%d,"id":%d,"apiVersionId":%d,"policyVersion":"v10"}`,
			configData,
			newPolicyTemplateID,
			policyID,
			apiInstanceID,
		)
	req, _ := client.newRequest("PATCH", updateAPIInstancePolicyURL, bytes.NewBuffer([]byte(updatePolicyPayload)))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		return fmt.Errorf("Failed to update API Instance Policies status code: %d\nResponse Message:%s",
			res.StatusCode, string(bodyBytes))
	}

	return nil
}

type _ConfigurationData ConfigurationData

func (t *ConfigurationData) UnmarshalJSON(b []byte) error {
	t2 := _ConfigurationData{}
	err := json.Unmarshal(b, &t2)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &(t2.AdditionalPolicyProperties))
	if err != nil {
		return err
	}

	typ := reflect.TypeOf(t2)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "" && jsonTag != "-" {
			delete(t2.AdditionalPolicyProperties, jsonTag)
		}
	}

	*t = ConfigurationData(t2)

	return nil
}
