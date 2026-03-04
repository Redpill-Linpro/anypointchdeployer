package cmd

import (
	"reflect"
	"testing"

	"github.com/Redpill-Linpro/anypointchdeployer/pkg/anypointclient"
)

func TestPolicyMatchingWithPointcuts(t *testing.T) {
	tests := []struct {
		name           string
		existingPolicy anypointclient.ApiPolicyResponse
		requestPolicy  anypointclient.ApiPolicyRequest
		shouldMatch    bool
	}{
		{
			name: "Same policy type with no pointcuts should match",
			existingPolicy: anypointclient.ApiPolicyResponse{
				Template: struct {
					GroupID      string `json:"groupId,omitempty"`
					AssetID      string `json:"assetId,omitempty"`
					AssetVersion string `json:"assetVersion,omitempty"`
				}{
					GroupID: "group1",
					AssetID: "asset1",
				},
				PointcutData: nil,
			},
			requestPolicy: anypointclient.ApiPolicyRequest{
				GroupID:      "group1",
				AssetID:      "asset1",
				PointcutData: nil,
			},
			shouldMatch: true,
		},
		{
			name: "Same policy type with identical pointcuts should match",
			existingPolicy: anypointclient.ApiPolicyResponse{
				Template: struct {
					GroupID      string `json:"groupId,omitempty"`
					AssetID      string `json:"assetId,omitempty"`
					AssetVersion string `json:"assetVersion,omitempty"`
				}{
					GroupID: "group1",
					AssetID: "client-authorization-policy",
				},
				PointcutData: []any{
					map[string]any{
						"methodRegex":      "GET",
						"uriTemplateRegex": "^/health$",
					},
				},
			},
			requestPolicy: anypointclient.ApiPolicyRequest{
				GroupID: "group1",
				AssetID: "client-authorization-policy",
				PointcutData: []any{
					map[string]any{
						"methodRegex":      "GET",
						"uriTemplateRegex": "^/health$",
					},
				},
			},
			shouldMatch: true,
		},
		{
			name: "Same policy type with different pointcuts should NOT match",
			existingPolicy: anypointclient.ApiPolicyResponse{
				Template: struct {
					GroupID      string `json:"groupId,omitempty"`
					AssetID      string `json:"assetId,omitempty"`
					AssetVersion string `json:"assetVersion,omitempty"`
				}{
					GroupID: "group1",
					AssetID: "client-authorization-policy",
				},
				PointcutData: []any{
					map[string]any{
						"methodRegex":      "GET",
						"uriTemplateRegex": "^/health$",
					},
				},
			},
			requestPolicy: anypointclient.ApiPolicyRequest{
				GroupID: "group1",
				AssetID: "client-authorization-policy",
				PointcutData: []any{
					map[string]any{
						"methodRegex":      "POST",
						"uriTemplateRegex": "^/api$",
					},
				},
			},
			shouldMatch: false,
		},
		{
			name: "Same policy type where one has pointcut and other doesn't should NOT match",
			existingPolicy: anypointclient.ApiPolicyResponse{
				Template: struct {
					GroupID      string `json:"groupId,omitempty"`
					AssetID      string `json:"assetId,omitempty"`
					AssetVersion string `json:"assetVersion,omitempty"`
				}{
					GroupID: "group1",
					AssetID: "client-authorization-policy",
				},
				PointcutData: nil,
			},
			requestPolicy: anypointclient.ApiPolicyRequest{
				GroupID: "group1",
				AssetID: "client-authorization-policy",
				PointcutData: []any{
					map[string]any{
						"methodRegex":      "GET",
						"uriTemplateRegex": "^/health$",
					},
				},
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is the matching logic we use in deployApiPolicy
			matches := tt.existingPolicy.Template.GroupID == tt.requestPolicy.GroupID &&
				tt.existingPolicy.Template.AssetID == tt.requestPolicy.AssetID &&
				reflect.DeepEqual(tt.existingPolicy.PointcutData, tt.requestPolicy.PointcutData)

			if matches != tt.shouldMatch {
				t.Errorf("Policy matching failed: expected match=%v, got match=%v", tt.shouldMatch, matches)
			}
		})
	}
}
