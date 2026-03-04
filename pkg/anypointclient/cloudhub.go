package anypointclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type CloudhubDeploymentResp struct {
	ID               string `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	CreationDate     int64  `json:"creationDate,omitempty"`
	LastModifiedDate int64  `json:"lastModifiedDate,omitempty"`
	Target           struct {
		Provider           string `json:"provider,omitempty"`
		TargetID           string `json:"targetId,omitempty"`
		DeploymentSettings struct {
			Clustered                           bool                  `json:"clustered,omitempty"`
			EnforceDeployingReplicasAcrossNodes bool                  `json:"enforceDeployingReplicasAcrossNodes,omitempty"`
			HTTP                                DeploymentHttpIngress `json:"http"`
			Outbound                            struct {
			} `json:"outbound"`
			Jvm struct {
			} `json:"jvm"`
			RuntimeVersion        string `json:"runtimeVersion,omitempty"`
			RuntimeReleaseChannel string `json:"runtimeReleaseChannel,omitempty"`
			Runtime               struct {
				Version        string `json:"version,omitempty"`
				ReleaseChannel string `json:"releaseChannel,omitempty"`
				Java           string `json:"java,omitempty"`
			} `json:"runtime"`
			UpdateStrategy          string `json:"updateStrategy,omitempty"`
			DisableAmLogForwarding  bool   `json:"disableAmLogForwarding,omitempty"`
			PersistentObjectStore   bool   `json:"persistentObjectStore,omitempty"`
			AnypointMonitoringScope string `json:"anypointMonitoringScope,omitempty"`
			Sidecars                struct {
				AnypointMonitoring struct {
					Image     string `json:"image,omitempty"`
					Resources struct {
						CPU struct {
							Limit    string `json:"limit,omitempty"`
							Reserved string `json:"reserved,omitempty"`
						} `json:"cpu"`
						Memory struct {
							Limit    string `json:"limit,omitempty"`
							Reserved string `json:"reserved,omitempty"`
						} `json:"memory"`
					} `json:"resources"`
				} `json:"anypoint-monitoring"`
			} `json:"sidecars"`
			GenerateDefaultPublicURL bool `json:"generateDefaultPublicUrl,omitempty"`
		} `json:"deploymentSettings"`
		Replicas int `json:"replicas,omitempty"`
	} `json:"target"`
	Status      string `json:"status,omitempty"`
	Application struct {
		Status       string `json:"status,omitempty"`
		DesiredState string `json:"desiredState,omitempty"`
		Ref          struct {
			GroupID    string `json:"groupId,omitempty"`
			ArtifactID string `json:"artifactId,omitempty"`
			Version    string `json:"version,omitempty"`
			Packaging  string `json:"packaging,omitempty"`
		} `json:"ref"`
		Configuration struct {
			MuleAgentApplicationPropertiesService struct {
				ApplicationName string            `json:"applicationName,omitempty"`
				Properties      map[string]string `json:"properties,omitempty"`
			} `json:"mule.agent.application.properties.service"`
			MuleAgentLoggingService struct {
				ArtifactName               string `json:"artifactName,omitempty"`
				ScopeLoggingConfigurations []any  `json:"scopeLoggingConfigurations,omitempty"`
			} `json:"mule.agent.logging.service"`
			MuleAgentScheduleService struct {
				Schedulers []Schedule `json:"schedulers,omitempty"`
			} `json:"mule.agent.scheduling.service"`
		} `json:"configuration"`
		VCores float32 `json:"vCores,omitempty"`
	} `json:"application"`
	DesiredVersion string `json:"desiredVersion,omitempty"`
	Replicas       []struct {
		ID                       string `json:"id,omitempty"`
		State                    string `json:"state,omitempty"`
		DeploymentLocation       string `json:"deploymentLocation,omitempty"`
		CurrentDeploymentVersion string `json:"currentDeploymentVersion,omitempty"`
		Reason                   string `json:"reason,omitempty"`
	} `json:"replicas,omitempty"`
	LastSuccessfulVersion string `json:"lastSuccessfulVersion,omitempty"`
}

type CloudhubDeploymentReq struct {
	Name   string   `json:"name"`
	Labels []string `json:"labels"`
	Target struct {
		Provider           string `json:"provider"`
		TargetID           string `json:"targetId"`
		DeploymentSettings struct {
			Clustered                           bool                  `json:"clustered"`
			EnforceDeployingReplicasAcrossNodes bool                  `json:"enforceDeployingReplicasAcrossNodes"`
			HTTP                                DeploymentHttpIngress `json:"http"`
			Jvm                                 struct {
			} `json:"jvm"`
			Outbound struct {
			} `json:"outbound"`
			RuntimeVersion        string `json:"runtimeVersion"`
			RuntimeReleaseChannel string `json:"runtimeReleaseChannel,omitempty"`
			Runtime               struct {
				Version        string `json:"version,omitempty"`
				ReleaseChannel string `json:"releaseChannel,omitempty"`
				Java           string `json:"java,omitempty"`
			} `json:"runtime"`
			UpdateStrategy           string `json:"updateStrategy"`
			DisableAmLogForwarding   bool   `json:"disableAmLogForwarding"`
			PersistentObjectStore    bool   `json:"persistentObjectStore"`
			GenerateDefaultPublicURL bool   `json:"generateDefaultPublicUrl"`
		} `json:"deploymentSettings"`
		Replicas int `json:"replicas"`
	} `json:"target"`
	Application struct {
		Ref struct {
			GroupID    string `json:"groupId"`
			ArtifactID string `json:"artifactId"`
			Version    string `json:"version"`
			Packaging  string `json:"packaging"`
		} `json:"ref"`
		Assets        []any  `json:"assets"`
		DesiredState  string `json:"desiredState"`
		Configuration struct {
			MuleAgentApplicationPropertiesService struct {
				ApplicationName  string            `json:"applicationName"`
				Properties       map[string]string `json:"properties"`
				SecureProperties map[string]string `json:"secureProperties"`
			} `json:"mule.agent.application.properties.service"`
			MuleAgentLoggingService struct {
				ScopeLoggingConfigurations []any `json:"scopeLoggingConfigurations"`
			} `json:"mule.agent.logging.service"`
			MuleAgentScheduleService struct {
				Schedulers []Schedule `json:"schedulers"`
			} `json:"mule.agent.scheduling.service"`
		} `json:"configuration"`
		Integrations struct {
			Services struct {
				ObjectStoreV2 struct {
					Enabled bool `json:"enabled"`
				} `json:"objectStoreV2"`
			} `json:"services"`
		} `json:"integrations"`
		VCores float32 `json:"vCores"`
	} `json:"application"`
}

type DeploymentHttpIngress struct {
	Inbound struct {
		PublicURL         string            `json:"publicUrl"`
		PathRewrite       string            `json:"pathRewrite"`
		LastMileSecurity  bool              `json:"lastMileSecurity"`
		ForwardSslSession bool              `json:"forwardSslSession"`
		InternalURL       string            `json:"internalUrl,omitempty"`
		Endpoints         []IngressEndpoint `json:"endpoints,omitempty"`
	} `json:"inbound"`
}

type IngressEndpoint struct {
	URL         string `json:"url"`
	PathRewrite string `json:"pathRewrite"`
	Access      string `json:"access"`
}

type Schedule struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	FlowName   string `json:"flowName"`
	Enabled    bool   `json:"enabled"`
	Expression string `json:"expression"`
	TimeZone   string `json:"timeZone"`
}
type ScheduleResp struct {
	Total int        `json:"total"`
	Items []Schedule `json:"items"`
}

type CloudhubDeploymentsResp struct {
	Total      int          `json:"total,omitempty"`
	Deloyments []Deployment `json:"items,omitempty"`
}

type Deployment struct {
	ID               string `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	CreationDate     int64  `json:"creationDate,omitempty"`
	LastModifiedDate int64  `json:"lastModifiedDate,omitempty"`
	Target           struct {
		Provider string `json:"provider,omitempty"`
		TargetID string `json:"targetId,omitempty"`
	} `json:"target"`
	Status      string `json:"status,omitempty"`
	Application struct {
		Status string `json:"status,omitempty"`
	} `json:"application"`
	CurrentRuntimeVersion        string `json:"currentRuntimeVersion,omitempty"`
	LastSuccessfulRuntimeVersion string `json:"lastSuccessfulRuntimeVersion,omitempty"`
}

func (client *AnypointClient) GetDeployments(environment Environment) ([]Deployment, error) {
	reqPath := fmt.Sprintf("/amc/application-manager/api/v2/organizations/%s/environments/%s/deployments", environment.OrganizationID, environment.ID)
	req, _ := client.newRequest("GET", reqPath, nil)

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, wrapError(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("call to Anypoint Platform returned %d", res.StatusCode)
	}

	var deploymentsResp CloudhubDeploymentsResp
	err = decodeResponseBody(res.Body, &deploymentsResp)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to process response")
	}

	return deploymentsResp.Deloyments, nil
}

func (client *AnypointClient) getDeloymentId(environment Environment, deploymentName string) (string, error) {
	deployments, err := client.GetDeployments(environment)
	if err != nil {
		return "", err
	}

	for _, deployment := range deployments {
		if deployment.Name == deploymentName {
			return deployment.ID, nil
		}
	}
	return "", nil
}

func (client *AnypointClient) GetDeployment(environment Environment, deploymentName string) (CloudhubDeploymentResp, error) {

	deploymentId, err := client.getDeloymentId(environment, deploymentName)
	if err != nil {
		return CloudhubDeploymentResp{}, err
	}
	if deploymentId == "" {
		return CloudhubDeploymentResp{}, nil
	}

	reqPath := fmt.Sprintf("/amc/application-manager/api/v2/organizations/%s/environments/%s/deployments/%s", environment.OrganizationID, environment.ID, deploymentId)
	req, err := client.newRequest("GET", reqPath, nil)
	if err != nil {
		return CloudhubDeploymentResp{}, errors.Wrap(err, "failed to create new request")
	}

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return CloudhubDeploymentResp{}, errors.Wrapf(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return handleErrorResponse(res)
	}

	var response CloudhubDeploymentResp

	err = decodeResponseBody(res.Body, &response)
	if err != nil {
		return CloudhubDeploymentResp{}, errors.Wrap(err, "Failed to process response")
	}
	return response, nil
}

/*--------------------*/

func handleErrorResponse(res *http.Response) (CloudhubDeploymentResp, error) {
	var response map[string]any
	err := decodeResponseBody(res.Body, &response)
	if err != nil {
		return CloudhubDeploymentResp{}, errors.Wrapf(err, "call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)
	}

	if res.StatusCode == http.StatusNotFound {
		return CloudhubDeploymentResp{}, nil
	}

	return CloudhubDeploymentResp{}, errors.Errorf("Call to Anypoint Platform returned %d : %s", res.StatusCode, response["message"])
}

func decodeResponseBody(body io.Reader, target any) error {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return errors.Wrap(err, "Failed to read response")
	}
	err = json.Unmarshal(bodyBytes, target)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal response")
	}

	return nil
}

/*---------------------------------------------*/

func (client *AnypointClient) DeleteDeployment(environment Environment, privateSpace PrivateSpace, deploymentID string) error {
	reqPath := fmt.Sprintf("/amc/application-manager/api/v2/organizations/%s/environments/%s/deployments/%s", environment.OrganizationID, environment.ID, deploymentID)
	req, _ := client.newRequest("DELETE", reqPath, nil)

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Failed to call Anypoint Platform")
	}
	if res.StatusCode != http.StatusNoContent {
		var response map[string]any
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)
		}
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)

		}
		return errors.Errorf("Call to Anypoint Platform returned %d : %s",
			res.StatusCode, response["message"])
	}
	return nil
}

func (client *AnypointClient) CreateDeployment(environment Environment, privateSpace PrivateSpace, deployment CloudhubDeploymentReq) (CloudhubDeploymentResp, error) {
	deployment.Target.TargetID = privateSpace.ID
	buffer := new(bytes.Buffer)

	// Remove leading ~ from version
	deployment.Target.DeploymentSettings.Runtime.Version = strings.TrimPrefix(deployment.Target.DeploymentSettings.Runtime.Version, "~")

	// If deployment.
	err := json.NewEncoder(buffer).Encode(deployment)

	if err != nil {
		errors.Wrapf(err, "failed to encode request")
	}
	reqPath := fmt.Sprintf("/amc/application-manager/api/v2/organizations/%s/environments/%s/deployments", environment.OrganizationID, environment.ID)

	req, _ := client.newRequest("POST", reqPath, buffer)

	req.Header.Add("Content-Type", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return CloudhubDeploymentResp{}, errors.Wrapf(err, "Failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		var response map[string]any
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return CloudhubDeploymentResp{}, errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)
		}
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return CloudhubDeploymentResp{}, errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)

		}
		return CloudhubDeploymentResp{}, errors.Errorf("Call to Anypoint Platform returned %d : %s",
			res.StatusCode, response["message"])
	}

	var response CloudhubDeploymentResp
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return CloudhubDeploymentResp{}, errors.Wrap(err, "Unable to decode response")
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return CloudhubDeploymentResp{}, errors.Wrap(err, "Unable to unmarshal response")
	}
	return response, nil
}

func (client *AnypointClient) UpdateDeployment(
	environment Environment,
	privateSpace PrivateSpace,
	deployment CloudhubDeploymentReq,
	deploymentID string) error {

	deployment.Target.TargetID = privateSpace.ID
	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(deployment)

	if err != nil {
		errors.Wrapf(err, "failed to encode request")
	}

	reqPath := fmt.Sprintf("/amc/application-manager/api/v2/organizations/%s/environments/%s/deployments/%s", environment.OrganizationID, environment.ID, deploymentID)

	req, _ := client.newRequest("PATCH", reqPath, buffer)

	req.Header.Add("Content-Type", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var response map[string]any
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)
		}
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return errors.Wrapf(err, "Call to Anypoint Platform returned %d. Failed to decode error response payload", res.StatusCode)

		}
		return errors.Errorf("Call to Anypoint Platform returned %d : %s",
			res.StatusCode, response["message"])
	}

	var response CloudhubDeploymentResp
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "Unable to decode response")
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return errors.Wrap(err, "Unable to unmarshal response")
	}
	return nil
}

func (client *AnypointClient) SchedulesDiffFromSourceCode(
	environment Environment,
	newDeployment CloudhubDeploymentReq,
	deploymentID string) error {

	reqPath := fmt.Sprintf("/amc/application-manager/api/v2/organizations/%s/environments/%s/deployments/%s/schedulers", environment.OrganizationID, environment.ID, deploymentID)
	req, _ := client.newRequest("GET", reqPath, nil)

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to call Anypoint Platform")
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.Errorf("call to Anypoint Platform returned %d", res.StatusCode)
	}

	var response ScheduleResp
	err = decodeResponseBody(res.Body, &response)
	if err != nil {
		return errors.Wrap(err, "Failed to process response")
	}

	err = compareSchedules(newDeployment.Application.Configuration.MuleAgentScheduleService.Schedulers, response.Items)
	if err != nil {
		return errors.Wrap(err, "Mismatch in Schedulers: The deployment configuration and the source code define different sets of schedulers or have different scheduler FlowName or Type")
	}
	return nil

}

func compareSchedules(slice1, slice2 []Schedule) error {
	if len(slice1) != len(slice2) {
		return fmt.Errorf("configuration has %d schedulers, but source code defines %d schedulers", len(slice1), len(slice2))
	}

	map1 := make(map[string]string)
	map2 := make(map[string]string)

	for _, obj := range slice1 {
		map1[obj.FlowName] = obj.Type
	}

	for _, obj := range slice2 {
		map2[obj.FlowName] = obj.Type
	}

	if !reflect.DeepEqual(map1, map2) {
		return fmt.Errorf("Deployment configuration:\n%v\nSource code:\n%v", map1, map2)
	}

	return nil
}

func (client *AnypointClient) UpdateScheduleNames(schedulers []Schedule) {
	for i := range schedulers {
		scheduler := &schedulers[i]
		scheduler.Name = "polling://" + scheduler.FlowName + "/"
	}
}

func wrapError(err error, message string) error {
	return errors.Wrap(err, message)
}
