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

const oldJson string = `
	{
		"A": "true",
		"B": "System",
		"C": "production environment",
		"D": "https://httpbin.org/anything",
		"chdeployer.application.gav": "someGroupId:someArtifactId:someVersion"
	}
	`

func TestPropertiesHasNotChanged(t *testing.T) {

	oldPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(oldJson), &oldPropertiesMap)

	newPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(oldJson), &newPropertiesMap)

	// Validate no changes
	if propertiesHasChanged(oldPropertiesMap, newPropertiesMap) {
		t.Errorf("No config diff, test should pass")
	}
}

func TestPropertiesOneValueHasChanged(t *testing.T) {
	oldPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(oldJson), &oldPropertiesMap)

	newPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(oldJson), &newPropertiesMap)

	// One value has changed in the new properties
	newPropertiesMap["A"] = "badVal"

	if resp := propertiesHasChanged(oldPropertiesMap, newPropertiesMap); resp != true {
		t.Errorf("Configs diff, evaluated to %v, should have evluated to false", resp)
	}
}

func TestPropertiesHasNotChangedWithSecureProperties(t *testing.T) {
	oldPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(oldJson), &oldPropertiesMap)

	newPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(oldJson), &newPropertiesMap)

	oldPropertiesMap["A"] = "*******"
	if propertiesHasChanged(oldPropertiesMap, newPropertiesMap) {
		t.Errorf("No config diff, since property A is secure (hidden), test should have passed")
	}
}

func TestPropertiesGAVHasChanged(t *testing.T) {
	oldPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(oldJson), &oldPropertiesMap)

	newPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(oldJson), &newPropertiesMap)
	newPropertiesMap["chdeployer.application.gav"] = "someGroupId:someArtifactId:someVersion2"

	if resp := propertiesHasChanged(oldPropertiesMap, newPropertiesMap); resp != true {
		t.Errorf("Configs diff new property added, evaluated to %v, should have evluated to true", resp)
	}

}

func TestPropertiesHasChangedWithOneNewProperty(t *testing.T) {
	oldPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(oldJson), &oldPropertiesMap)

	newPropertyJson := `
	{
		"A": "true",
		"B": "System",
		"C": "production environment",
		"D": "https://httpbin.org/anything",
		"E": "443",
		"chdeployer.application.gav": "someGroupId:someArtifactId:someVersion"
	}
	`

	newPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(newPropertyJson), &newPropertiesMap)
	if resp := propertiesHasChanged(oldPropertiesMap, newPropertiesMap); resp != true {
		t.Errorf("Configs diff new property added, evaluated to %v, should have evluated to true", resp)
	}

}

func TestPropertiesHasChangedWithOneRemovedProperty(t *testing.T) {
	oldPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(oldJson), &oldPropertiesMap)

	removedPropertyJson := `
	{
		"A": "true",
		"B": "System",
		"D": "https://httpbin.org/anything",
		"chdeployer.application.gav": "someGroupId:someArtifactId:someVersion"
	}
	`

	newPropertiesMap := make(map[string]string)
	json.Unmarshal([]byte(removedPropertyJson), &newPropertiesMap)
	if resp := propertiesHasChanged(oldPropertiesMap, newPropertiesMap); resp != true {
		t.Errorf("Configs diff property C removed, evaluated to %v, should have evluated to true", resp)
	}
}
