package anypointclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type CloudhubApplicationResponse struct {
	VersionID         string            `json:"versionId"`
	Domain            string            `json:"domain"`
	FullDomain        string            `json:"fullDomain"`
	Properties        map[string]string `json:"properties"`
	PropertiesOptions map[string]string `json:"propertiesOptions"`
	Status            string            `json:"status"`
	Workers           struct {
		Type struct {
			Name   string  `json:"name"`
			Weight float64 `json:"weight"`
			CPU    string  `json:"cpu"`
			Memory string  `json:"memory"`
		} `json:"type"`
		Amount              int     `json:"amount"`
		RemainingOrgWorkers float64 `json:"remainingOrgWorkers"`
		TotalOrgWorkers     float64 `json:"totalOrgWorkers"`
	} `json:"workers"`
	LastUpdateTime int64  `json:"lastUpdateTime"`
	FileName       string `json:"fileName"`
	MuleVersion    struct {
		Version          string `json:"version"`
		UpdateID         string `json:"updateId"`
		LatestUpdateID   string `json:"latestUpdateId"`
		EndOfSupportDate int64  `json:"endOfSupportDate"`
	} `json:"muleVersion"`
	Region                            string `json:"region"`
	PersistentQueues                  bool   `json:"persistentQueues"`
	PersistentQueuesEncryptionEnabled bool   `json:"persistentQueuesEncryptionEnabled"`
	PersistentQueuesEncrypted         bool   `json:"persistentQueuesEncrypted"`
	MonitoringEnabled                 bool   `json:"monitoringEnabled"`
	MonitoringAutoRestart             bool   `json:"monitoringAutoRestart"`
	StaticIPsEnabled                  bool   `json:"staticIPsEnabled"`
	HasFile                           bool   `json:"hasFile"`
	SecureDataGatewayEnabled          bool   `json:"secureDataGatewayEnabled"`
	LoggingNgEnabled                  bool   `json:"loggingNgEnabled"`
	LoggingCustomLog4JEnabled         bool   `json:"loggingCustomLog4JEnabled"`
	CloudObjectStoreRegion            string `json:"cloudObjectStoreRegion"`
	InsightsReplayDataRegion          string `json:"insightsReplayDataRegion"`
	IsDeploymentWaiting               bool   `json:"isDeploymentWaiting"`
	DeploymentGroup                   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"deploymentGroup"`
	UpdateRuntimeConfig bool `json:"updateRuntimeConfig"`
	TrackingSettings    struct {
		TrackingLevel string `json:"trackingLevel"`
	} `json:"trackingSettings"`
	LogLevels   []interface{} `json:"logLevels"`
	IPAddresses []interface{} `json:"ipAddresses"`
}

type CloudhubApplicationRequest struct {
	ApplicationInfo struct {
		FileName    string `json:"fileName,omitempty"`
		MuleVersion struct {
			Version string `json:"version,omitempty"`
		} `json:"muleVersion,omitempty"`
		Properties       map[string]string `json:"properties,omitempty"`
		LogLevels        []interface{}     `json:"logLevels,omitempty"`
		TrackingSettings struct {
			TrackingLevel string `json:"trackingLevel,omitempty"`
		} `json:"trackingSettings,omitempty"`
		DeploymentGroup           interface{} `json:"deploymentGroup,omitempty"`
		MonitoringEnabled         bool        `json:"monitoringEnabled,omitempty"`
		MonitoringAutoRestart     bool        `json:"monitoringAutoRestart,omitempty"`
		PersistentQueues          bool        `json:"persistentQueues,omitempty"`
		PersistentQueuesEncrypted bool        `json:"persistentQueuesEncrypted,omitempty"`
		Workers                   struct {
			Amount int `json:"amount,omitempty"`
			Type   struct {
				Name   string  `json:"name,omitempty"`
				Weight float64 `json:"weight,omitempty"`
				CPU    string  `json:"cpu,omitempty"`
				Memory string  `json:"memory,omitempty"`
			} `json:"type,omitempty"`
		} `json:"workers,omitempty"`
		ObjectStoreV1             bool   `json:"objectStoreV1,omitempty"`
		LoggingNgEnabled          bool   `json:"loggingNgEnabled,omitempty"`
		LoggingCustomLog4JEnabled bool   `json:"loggingCustomLog4JEnabled,omitempty"`
		StaticIPsEnabled          bool   `json:"staticIPsEnabled,omitempty"`
		Domain                    string `json:"domain,omitempty"`
	} `json:"applicationInfo,omitempty"`
	ApplicationSource struct {
		Source         string `json:"source,omitempty"`
		GroupID        string `json:"groupId,omitempty"`
		ArtifactID     string `json:"artifactId,omitempty"`
		Version        string `json:"version,omitempty"`
		OrganizationID string `json:"organizationId,omitempty"`
	} `json:"applicationSource,omitempty"`
	AutoStart bool `json:"autoStart,omitempty"`
}

func (client *AnypointClient) GetApplications(environment Environment) ([]CloudhubApplicationResponse, error) {
	req, _ := client.newRequest("GET", "/cloudhub/api/v2/applications", nil)
	// Set X-ANYPNT-ENV-ID and X-ANYPNT-ORG-ID and possibly also X-CH-SuppressBasicAuth: 1
	req.Header.Add("X-ANYPNT-ENV-ID", environment.ID)
	req.Header.Add("X-ANYPNT-ORG-ID", environment.OrganizationID)

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var response map[string]interface{}
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil,
				errors.Wrapf(err, "call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)
		}
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return nil,
				errors.Wrapf(err, "call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)

		}
		return nil, Errorf("Call to Anypoint Platform returned %d : %s",
			res.StatusCode, response["message"])
	}

	var response []CloudhubApplicationResponse
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed do read response")
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed do unmarshal response")
	}
	return response, nil
}

func (client *AnypointClient) GetApplication(environment Environment, applicationname string) (CloudhubApplicationResponse, error) {
	req, _ := client.newRequest("GET",
		fmt.Sprintf("/cloudhub/api/v2/applications/%s", applicationname), nil)
	// Set X-ANYPNT-ENV-ID and X-ANYPNT-ORG-ID and possibly also X-CH-SuppressBasicAuth: 1
	req.Header.Add("X-ANYPNT-ENV-ID", environment.ID)
	req.Header.Add("X-ANYPNT-ORG-ID", environment.OrganizationID)

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return CloudhubApplicationResponse{}, errors.Wrapf(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var response map[string]interface{}
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return CloudhubApplicationResponse{},
				errors.Wrapf(err, "call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)
		}
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return CloudhubApplicationResponse{},
				errors.Wrapf(err, "call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)

		}
		if res.StatusCode == http.StatusNotFound {
			return CloudhubApplicationResponse{}, nil
		} else {
			return CloudhubApplicationResponse{}, Errorf("Call to Anypoint Platform returned %d : %s",
				res.StatusCode, response["message"])
		}
	}

	var response CloudhubApplicationResponse
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return CloudhubApplicationResponse{},
			errors.Wrapf(err, "Failed do read response")
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return CloudhubApplicationResponse{},
			errors.Wrapf(err, "Failed do unmarshal response")
	}
	return response, nil
}

func (client *AnypointClient) CreateApplication(environment Environment, application CloudhubApplicationRequest) error {
	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(application)

	if err != nil {
		errors.Wrapf(err, "failed to encode request")
	}

	req, _ := client.newRequest("POST", "/cloudhub/api/v2/applications", buffer)
	// Set X-ANYPNT-ENV-ID and X-ANYPNT-ORG-ID and possibly also X-CH-SuppressBasicAuth: 1
	req.Header.Add("X-ANYPNT-ENV-ID", environment.ID)
	req.Header.Add("X-ANYPNT-ORG-ID", environment.OrganizationID)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Failed to call Anypoint Platform: %w")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var response map[string]interface{}
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)
		}
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)

		}
		return Errorf("Call to Anypoint Platform returned %d : %s",
			res.StatusCode, response["message"])
	}

	var response CloudhubApplicationResponse
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "Unable to decode response")
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return errors.Wrap(err, "Unable to unmarshal response")
	}
	return nil
}

func (client *AnypointClient) UpdateApplication(environment Environment, application CloudhubApplicationRequest) error {
	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(application)

	if err != nil {
		errors.Wrapf(err, "failed to encode request")
	}

	req, _ := client.newRequest("PUT",
		fmt.Sprintf("/cloudhub/api/v2/applications/%s", application.ApplicationInfo.Domain),
		buffer)
	// Set X-ANYPNT-ENV-ID and X-ANYPNT-ORG-ID and possibly also X-CH-SuppressBasicAuth: 1
	req.Header.Add("X-ANYPNT-ENV-ID", environment.ID)
	req.Header.Add("X-ANYPNT-ORG-ID", environment.OrganizationID)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Failed to call Anypoint Platform: %w")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var response map[string]interface{}
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)
		}
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)

		}
		return Errorf("Call to Anypoint Platform returned %d : %s",
			res.StatusCode, response["message"])
	}

	var response CloudhubApplicationResponse
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "Unable to decode response")
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return errors.Wrap(err, "Unable to unmarshal response")
	}
	return nil
}
