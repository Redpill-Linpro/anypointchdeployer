package anypointclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PrivateSpace struct {
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name,omitempty"`
	OrganizationID     string `json:"organizationId,omitempty"`
	RootOrganizationID string `json:"rootOrganizationId,omitempty"`
	Region             string `json:"region,omitempty"`
	Status             string `json:"status,omitempty"`
	Connections        struct {
	} `json:"connections"`
	IsSpaceUpgradeAvailable bool `json:"isSpaceUpgradeAvailable,omitempty"`
}

type PrivateSpacesResponse struct {
	PrivateSpaces []PrivateSpace `json:"content,omitempty"`
	Pageable      struct {
		Sort struct {
			Empty    bool `json:"empty,omitempty"`
			Sorted   bool `json:"sorted,omitempty"`
			Unsorted bool `json:"unsorted,omitempty"`
		} `json:"sort"`
		Offset     int  `json:"offset,omitempty"`
		PageNumber int  `json:"pageNumber,omitempty"`
		PageSize   int  `json:"pageSize,omitempty"`
		Paged      bool `json:"paged,omitempty"`
		Unpaged    bool `json:"unpaged,omitempty"`
	} `json:"pageable"`
	Last          bool `json:"last,omitempty"`
	TotalPages    int  `json:"totalPages,omitempty"`
	TotalElements int  `json:"totalElements,omitempty"`
	Size          int  `json:"size,omitempty"`
	Number        int  `json:"number,omitempty"`
	Sort          struct {
		Empty    bool `json:"empty,omitempty"`
		Sorted   bool `json:"sorted,omitempty"`
		Unsorted bool `json:"unsorted,omitempty"`
	} `json:"sort"`
	First            bool `json:"first,omitempty"`
	NumberOfElements int  `json:"numberOfElements,omitempty"`
	Empty            bool `json:"empty,omitempty"`
}

/*
ResolveEnvironment will resolve, in the given organization, an Environment by name.
*/
func (client *AnypointClient) ResolvePrivateSpace(organization Organization, privateSpaceName string) (PrivateSpace, error) {
	privateSpaceResponse := new(PrivateSpacesResponse)

	reqPath := fmt.Sprintf("runtimefabric/api/organizations/%s/privatespaces", organization.ID)
	res, err := client.newGetRequest(reqPath)
	if err != nil {
		return PrivateSpace{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return PrivateSpace{}, err
		}
		err = json.Unmarshal(bodyBytes, &privateSpaceResponse)
		if err != nil {
			return PrivateSpace{}, err
		}
	}

	for _, privateSpace := range privateSpaceResponse.PrivateSpaces {
		if privateSpace.Name == privateSpaceName {
			return privateSpace, nil
		}
	}
	return PrivateSpace{}, fmt.Errorf("failed to find environment named %s in organization %s", privateSpaceName, organization.Name)
}
