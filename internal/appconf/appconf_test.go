package appconf_test

import (
	"embed"
	"encoding/json"
	"strings"

	"github.com/Redpill-Linpro/anypointchdeployer/internal/appconf"
	"github.com/Redpill-Linpro/anypointchdeployer/pkg/anypointclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Add embedded filesystem with test resources
//
//go:embed resources
var testresources embed.FS

var _ = Describe("Appconf", func() {
	It("should be able to read simple-app", func() {

		// Read the file simple-app-v1.json as a string from the embedded resources testresources
		text, err := testresources.ReadFile("resources/simple-app-v1.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var deployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(text))).Decode(&deployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Validate the decoded deployment definition
		Ω(deployment.Name).Should(Equal("simple-app"), "Name")
	})

	It("detect change of application version", func() {

		// Read the file simple-app-current.json as a string from the embedded resources testresources
		responsetext, err := testresources.ReadFile("resources/simple-app-current.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read the file simple-app-v1.json as a string from the embedded resources testresources
		requesttext, err := testresources.ReadFile("resources/simple-app-v1.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		Ω(changed).Should(BeTrue(), "Should detect change")
	})

	It("detect change of runtime version", func() {

		// Read the file simple-app-current.json as a string from the embedded resources testresources
		responsetext, err := testresources.ReadFile("resources/simple-app-current.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read the file simple-app-v2.json as a string from the embedded resources testresources
		requesttext, err := testresources.ReadFile("resources/simple-app-v2.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		Ω(changed).Should(BeTrue(), "Should detect change")
	})

	It("detect change of runtime version (old syntax)", func() {

		// Read the file simple-app-current.json as a string from the embedded resources testresources
		responsetext, err := testresources.ReadFile("resources/simple-app-current.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read the file simple-app-v2-old-runtimeversion.json as a string from the embedded resources testresources
		requesttext, err := testresources.ReadFile("resources/simple-app-v2-old-runtimeversion.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Update the deployment to match latest schema version
		desiredDeployment, err = appconf.UpdateDeploymentToLatestSchema(desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		Ω(changed).Should(BeTrue(), "Should detect change")
	})

	It("testing tilde range", func() {

		// Read the file simple-app-current.json as a string from the embedded resources testresources
		responsetext, err := testresources.ReadFile("resources/simple-app-current.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read the file simple-app-v2-tilderange.json as a string from the embedded resources testresources
		requesttext, err := testresources.ReadFile("resources/simple-app-v2-tilderange.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		Ω(changed).Should(BeFalse(), "Should not detect change")
	})

	It("detect multiple changes", func() {

		// Read the file simple-app-current.json as a string from the embedded resources testresources
		responsetext, err := testresources.ReadFile("resources/simple-app-current.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read the file simple-app-v3.json as a string from the embedded resources testresources
		requesttext, err := testresources.ReadFile("resources/simple-app-v3.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		Ω(changed).Should(BeTrue(), "Should detect change")
	})

	It("property change", func() {

		// Read the file simple-app-current.json as a string from the embedded resources testresources
		responsetext, err := testresources.ReadFile("resources/simple-app-current.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read the file simple-app-v4.json as a string from the embedded resources testresources
		requesttext, err := testresources.ReadFile("resources/simple-app-v4.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		Ω(changed).Should(BeTrue(), "Should detect change")
	})

	It("publicUrl change", func() {

		// Read the file simple-app-current.json as a string from the embedded resources testresources
		responsetext, err := testresources.ReadFile("resources/simple-app-current.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read the file simple-app-v2-change-publicUrl.json as a string from the embedded resources testresources
		requesttext, err := testresources.ReadFile("resources/simple-app-v2-change-publicUrl.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		Ω(changed).Should(BeTrue(), "Should detect change")
	})

	It("should handle deployments with endpoints configuration", func() {
		// Read the file simple-app-with-endpoints.json
		requesttext, err := testresources.ReadFile("resources/simple-app-with-endpoints.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Create a new deployment definition from the string
		var deployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&deployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Validate the decoded deployment definition
		Ω(deployment.Name).Should(Equal("simple-app"), "Name")
		Ω(deployment.Target.DeploymentSettings.HTTP.Inbound.Endpoints).Should(HaveLen(3), "Should have 3 endpoints")
		Ω(deployment.Target.DeploymentSettings.HTTP.Inbound.Endpoints[0].URL).Should(Equal("https://api.example.com/my-api/v1"))
		Ω(deployment.Target.DeploymentSettings.HTTP.Inbound.Endpoints[0].Access).Should(Equal("external"))
		Ω(deployment.Target.DeploymentSettings.HTTP.Inbound.InternalURL).Should(Equal("https://my-api-dhmnv8.internal-e676y6.deu-c1.eu1.cloudhub.io"))
	})

	It("should detect endpoint changes", func() {
		// Read the file with original endpoints
		responsetext, err := testresources.ReadFile("resources/simple-app-with-endpoints.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read the file with changed endpoints
		requesttext, err := testresources.ReadFile("resources/simple-app-with-endpoints-changed.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		Ω(changed).Should(BeTrue(), "Should detect endpoint URL change")
	})

	It("should handle backwards compatibility when endpoints are not specified", func() {
		// Read a file without endpoints (old format)
		responsetext, err := testresources.ReadFile("resources/simple-app-v1.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read another file without endpoints (same config, no endpoints)
		requesttext, err := testresources.ReadFile("resources/simple-app-v1.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Should work without endpoints (backwards compatibility)
		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		// The test should not panic and should handle the case gracefully
		Ω(changed).Should(BeFalse(), "Should not detect changes when they are identical without endpoints")
	})

	It("should not trigger redeploy when config has no endpoints but deployment has endpoints", func() {
		// Read a deployment response WITH endpoints (as would come from Mule API)
		responsetext, err := testresources.ReadFile("resources/simple-app-with-endpoints.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read a config WITHOUT endpoints that matches the publicUrl from deployment
		requesttext, err := testresources.ReadFile("resources/simple-app-matching-publicurl.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Should NOT trigger a change just because deployment has endpoints that config doesn't specify
		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		Ω(changed).Should(BeFalse(), "Should not detect changes when config has no endpoints but matches publicUrl")
	})

	It("should detect publicUrl changes even when endpoints are present", func() {
		// Read a deployment response WITH endpoints
		responsetext, err := testresources.ReadFile("resources/simple-app-with-endpoints.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		var currentDeployment anypointclient.CloudhubDeploymentResp
		err = json.NewDecoder(strings.NewReader(string(responsetext))).Decode(&currentDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Read the same config with endpoints
		requesttext, err := testresources.ReadFile("resources/simple-app-with-endpoints.json")
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		var desiredDeployment anypointclient.CloudhubDeploymentReq
		err = json.NewDecoder(strings.NewReader(string(requesttext))).Decode(&desiredDeployment)
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)

		// Change the publicUrl to test that it's always checked
		desiredDeployment.Target.DeploymentSettings.HTTP.Inbound.PublicURL = "https://different-url.example.com"

		// Should detect the publicUrl change even though endpoints match
		_, changed := appconf.PrepareDeploymentAndCheckChanges(desiredDeployment, currentDeployment)
		Ω(changed).Should(BeTrue(), "Should detect publicUrl change even with endpoints present")
	})
})
