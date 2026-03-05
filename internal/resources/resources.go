package resources

import "github.com/Redpill-Linpro/anypointchdeployer/pkg/anypointclient"

type BaseResource struct {
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

type ApplicationV1 struct {
	BaseResource
	Spec anypointclient.CloudhubDeploymentReq `json:"spec"`
}

type ApiPoliciesV1 struct {
	BaseResource
	Spec struct {
		ApiInstanceID string                            `json:"apiInstanceId"`
		Policies      []anypointclient.ApiPolicyRequest `json:"policy"`
	} `json:"spec"`
}

// MqExchangeWithBindings extends MqExchange with bindings configuration
type MqExchangeWithBindings struct {
	anypointclient.MqExchange
	Bindings []anypointclient.MqBinding `json:"bindings,omitempty"`
}

type MqDestinationsV1 struct {
	BaseResource
	Spec struct {
		Queues    []anypointclient.MqQueue `json:"queues,omitempty"`
		Exchanges []MqExchangeWithBindings `json:"exchanges,omitempty"`
	} `json:"spec"`
}
