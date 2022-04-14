package appconf

import (
	"encoding/json"
	"testing"

	"github.com/Redpill-Linpro/anypointchdeployer/pkg/anypointclient"
)

func TestAddApplicationGav(t *testing.T) {
	var app anypointclient.CloudhubApplicationRequest

	app.ApplicationInfo.Properties = make(map[string]string)

	app.ApplicationSource.GroupID = "someGroupId"
	app.ApplicationSource.ArtifactID = "someArtifactId"
	app.ApplicationSource.Version = "someVersion"

	AddApplicationGav(app)

	if app.ApplicationInfo.Properties["chdeployer.application.gav"] != "someGroupId:someArtifactId:someVersion" {
		t.Errorf("\nExpected: someGroupId:someArtifactId:someVersion\nReceived: %s", app.ApplicationInfo.Properties["chdeployer.application.gav"])
	}
}

func TestPropertiesHasChanged(t *testing.T) {
	type properties struct {
		A string
		B string
		C string
		D string
	}

	old := properties{
		A: "true",
		B: "System",
		C: "production environment",
		D: "https://httpbin.org/anything",
	}

	new := properties{
		A: "true",
		B: "System",
		C: "production environment",
		D: "https://httpbin.org/anything",
	}

	oldPropertiesMap := make(map[string]string)
	applicationData, _ := json.Marshal(old)
	json.Unmarshal(applicationData, &oldPropertiesMap)

	newPropertiesMap := make(map[string]string)
	newApplicationData, _ := json.Marshal(new)
	json.Unmarshal(newApplicationData, &newPropertiesMap)

	if propertiesHasChanged(oldPropertiesMap, newPropertiesMap) {
		t.Errorf("No config diff, test should pass")
	}
	newPropertiesMap["A"] = "badVal"

	if resp := propertiesHasChanged(oldPropertiesMap, newPropertiesMap); resp != true {
		t.Errorf("Configs diff, evaluated to %v, should have evluated to false", resp)
	}

}
