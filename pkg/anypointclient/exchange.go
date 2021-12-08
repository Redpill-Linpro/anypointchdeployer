package anypointclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/xerrors"
)

/*

 */

type ExchangeAsset struct {
	GroupID      string      `json:"groupId"`
	AssetID      string      `json:"assetId"`
	Version      string      `json:"version"`
	MinorVersion string      `json:"minorVersion"`
	VersionGroup string      `json:"versionGroup"`
	Description  string      `json:"description"`
	IsPublic     bool        `json:"isPublic"`
	Name         string      `json:"name"`
	ContactName  interface{} `json:"contactName"`
	ContactEmail interface{} `json:"contactEmail"`
	Type         string      `json:"type"`
	IsSnapshot   bool        `json:"isSnapshot"`
	Status       string      `json:"status"`
	ExternalFile struct {
		URL interface{} `json:"url"`
	} `json:"externalFile"`
	CreatedDate    time.Time     `json:"createdDate"`
	UpdatedDate    time.Time     `json:"updatedDate"`
	MinMuleVersion interface{}   `json:"minMuleVersion"`
	Labels         []string      `json:"labels"`
	Categories     []interface{} `json:"categories"`
	Files          []struct {
		Classifier   string    `json:"classifier"`
		Packaging    string    `json:"packaging"`
		ExternalLink string    `json:"externalLink"`
		CreatedDate  time.Time `json:"createdDate"`
		Md5          string    `json:"md5"`
		Sha1         string    `json:"sha1"`
		MainFile     string    `json:"mainFile"`
		IsGenerated  bool      `json:"isGenerated"`
	} `json:"files"`
	CustomFields  []interface{} `json:"customFields"`
	Rating        int           `json:"rating"`
	NumberOfRates int           `json:"numberOfRates"`
	CreatedBy     struct {
		ID        string `json:"id"`
		UserName  string `json:"userName"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"createdBy"`
	Instances []struct {
		VersionGroup                string      `json:"versionGroup"`
		OrganizationID              string      `json:"organizationId"`
		ID                          string      `json:"id"`
		GroupID                     string      `json:"groupId"`
		AssetID                     string      `json:"assetId"`
		Version                     string      `json:"version"`
		MinorVersion                string      `json:"minorVersion"`
		ProductAPIVersion           string      `json:"productAPIVersion"`
		EnvironmentID               interface{} `json:"environmentId"`
		ProviderID                  interface{} `json:"providerId"`
		EndpointURI                 string      `json:"endpointUri"`
		Name                        string      `json:"name"`
		IsPublic                    bool        `json:"isPublic"`
		Type                        string      `json:"type"`
		Deprecated                  interface{} `json:"deprecated"`
		Fullname                    string      `json:"fullname"`
		AssetName                   string      `json:"assetName"`
		EnvironmentName             string      `json:"environmentName,omitempty"`
		EnvironmentOrganizationName string      `json:"environmentOrganizationName,omitempty"`
	} `json:"instances"`
	Dependencies []struct {
		Organization struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organization"`
		GroupID      string   `json:"groupId"`
		AssetID      string   `json:"assetId"`
		Version      string   `json:"version"`
		MinorVersion string   `json:"minorVersion"`
		VersionGroup string   `json:"versionGroup"`
		Name         string   `json:"name"`
		Type         string   `json:"type"`
		Permissions  []string `json:"permissions"`
	} `json:"dependencies"`
	Organization struct {
		ID                     string        `json:"id"`
		Name                   string        `json:"name"`
		ParentOrganizationIds  []interface{} `json:"parentOrganizationIds"`
		SubOrganizationIds     []string      `json:"subOrganizationIds"`
		TenantOrganizationIds  []interface{} `json:"tenantOrganizationIds"`
		IsMaster               bool          `json:"isMaster"`
		Domain                 string        `json:"domain"`
		IsMulesoftOrganization bool          `json:"isMulesoftOrganization"`
	} `json:"organization"`
	ID         string      `json:"id"`
	Icon       interface{} `json:"icon"`
	CreatedAt  time.Time   `json:"createdAt"`
	ModifiedAt time.Time   `json:"modifiedAt"`
}

/*
GetExchangeAssets retrieves assets from exchange
*/
func (client *AnypointClient) GetExchangeAssets(orgId string, offset int, limit int) (*[]ExchangeAsset, error) {
	req, _ := client.newRequest("GET", "exchange/api/v2/assets", nil)
	// curl 'https://anypoint.mulesoft.com/exchange/api/v2/assets?search=&&domain=&&masterOrganizationId=xxx&offset=20&limit=20&sharedWithMe=&includeSnapshots=true'  -H 'authorization: bearer xxxxx'
	q := req.URL.Query()
	q.Add("search", "")
	/*
		types=api-group
		&types=connector
		&types=custom
		&types=example
		&types=extension
		&types=http-api
		&types=policy
		&types=raml-fragment
		&types=rest-api
		&types=soap-api
		&types=template
	*/
	q.Add("types", "http-api")
	q.Add("types", "soap-api")
	q.Add("types", "rest-api")
	q.Add("domain", "")
	q.Add("masterOrganizationId", orgId)
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))
	q.Add("sharedWithMe", "")
	q.Add("includeSnapshots", "true")
	req.URL.RawQuery = q.Encode()

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var response []ExchangeAsset
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
	return &response, nil
}

// curl 'https://anypoint.mulesoft.com/exchange/api/v2/assets/xxxx/api' -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0' -H 'Accept: application/json' -H 'Accept-Language: en,en-US;q=0.7,sv;q=0.3' --compressed 
func (client *AnypointClient) GetExchangeAssetsDetails(orgId string, assetId string) (*ExchangeAsset, error) {
	req, _ := client.newRequest("GET",
		fmt.Sprintf("exchange/api/v2/assets/%s/%s", orgId, assetId),
		nil)
	// curl 'https://eu1.anypoint.mulesoft.com/exchange/api/v2/assets/%s/%s'

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var response ExchangeAsset
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
	return &response, nil
}

// Update
// curl 'https://anypoint.mulesoft.com/exchange/api/v2/assets/xxxxxxxxxxxx/api/versionGroups/v1/instances/managed/1122344'
// -X PATCH  -H 'authorization: bearer xxxxxxxx' -H 'content-type: application/json'
// --data-raw '{"name":"v1:25252525","endpointUri":"https://api.example.com/api/v2/","isPublic":false}'
func (client *AnypointClient) UpdateExchangeApiManagedInstanceUrl(orgId string, assetId string, versionGroup string, instanceId string, newURL string) error {
	updateInstancePayload :=
		fmt.Sprintf(
			`{"name":"%s:%s","endpointUri":"%s"}`,
			versionGroup,
			instanceId,
			newURL,
		)

	req, _ := client.newRequest("PATCH",
		fmt.Sprintf("exchange/api/v2/assets/%s/%s/versionGroups/%s/instances/managed/%s", orgId, assetId, versionGroup, instanceId),
		bytes.NewBuffer([]byte(updateInstancePayload)))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return xerrors.Errorf("Failed to update Exchange instance: %+v\n\t%+v", req, res)
	}
	return nil
}
