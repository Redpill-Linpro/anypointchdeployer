package appconf

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/Redpill-Linpro/anypointchdeployer/pkg/anypointclient"
	"github.com/spf13/viper"
)

func UpdateDeploymentToLatestSchema(newDeployment anypointclient.CloudhubDeploymentReq) (anypointclient.CloudhubDeploymentReq, error) {
	// TODO: This shallow copy should be a deep copy
	var updatedDeployment anypointclient.CloudhubDeploymentReq = newDeployment

	// If newDeployment.Target.DeploymentSettings.Runtime is empty check if we can use newDeployment.Target.DeploymentSettings.RuntimeVersion instead
	if newDeployment.Target.DeploymentSettings.Runtime.Java == "" || newDeployment.Target.DeploymentSettings.Runtime.ReleaseChannel == "" || newDeployment.Target.DeploymentSettings.Runtime.Version == "" {
		updatedDeployment.Target.DeploymentSettings.Runtime = struct {
			Version        string "json:\"version,omitempty\""
			ReleaseChannel string "json:\"releaseChannel,omitempty\""
			Java           string "json:\"java,omitempty\""
		}{
			Version:        newDeployment.Target.DeploymentSettings.RuntimeVersion,
			ReleaseChannel: newDeployment.Target.DeploymentSettings.RuntimeReleaseChannel,
			Java:           "",
		}
		updatedDeployment.Target.DeploymentSettings.RuntimeVersion = ""
	} else {
		updatedDeployment.Target.DeploymentSettings.RuntimeVersion = ""
	}
	return updatedDeployment, nil
}

func PrepareDeploymentAndCheckChanges(
	newDeployment anypointclient.CloudhubDeploymentReq,
	deployment anypointclient.CloudhubDeploymentResp) (anypointclient.CloudhubDeploymentReq, bool) {

	// Clone newDeployment to a new struct called updatedDeployment
	updatedDeployment := newDeployment

	changed := false

	if newDeployment.Name != deployment.Name {
		fmt.Println("Name changed from", deployment.Name, "to", newDeployment.Name, "for deployment", deployment.Name)
		changed = true
	}

	if ingressUpdated(newDeployment.Target.DeploymentSettings.HTTP, deployment.Target.DeploymentSettings.HTTP) {
		changed = true
	}

	if newDeployment.Target.DeploymentSettings.Jvm != deployment.Target.DeploymentSettings.Jvm {
		fmt.Println("Jvm settings changed from", deployment.Target.DeploymentSettings.Jvm, "to", newDeployment.Target.DeploymentSettings.Jvm, "for deployment", deployment.Name)
		changed = true
	}

	// Otherwise use the runtime struct
	if runtimeVersionUpdated(newDeployment.Target.DeploymentSettings.Runtime.Version, deployment.Target.DeploymentSettings.Runtime.Version) {
		// Remove any tilde from the new version
		updatedDeployment.Target.DeploymentSettings.Runtime.Version = strings.TrimPrefix(newDeployment.Target.DeploymentSettings.Runtime.Version, "~")
		fmt.Println("Runtime.Version changed from", deployment.Target.DeploymentSettings.Runtime.Version, "to", newDeployment.Target.DeploymentSettings.Runtime.Version, "for deployment", deployment.Name)
		changed = true
	} else {
		// Use the existing runtime version to make sure that we do not change version if the new version was a tilde range
		updatedDeployment.Target.DeploymentSettings.Runtime.Version = deployment.Target.DeploymentSettings.Runtime.Version
	}
	// Also check if the Java version has changed
	if newDeployment.Target.DeploymentSettings.Runtime.Java != deployment.Target.DeploymentSettings.Runtime.Java {
		fmt.Println("Runtime.Java changed from", deployment.Target.DeploymentSettings.Runtime.Java, "to", newDeployment.Target.DeploymentSettings.Runtime.Java, "for deployment", deployment.Name)
		changed = true
	}

	if newDeployment.Target.DeploymentSettings.DisableAmLogForwarding != deployment.Target.DeploymentSettings.DisableAmLogForwarding {
		fmt.Println("DisableAmLogForwarding changed from", deployment.Target.DeploymentSettings.DisableAmLogForwarding, "to", newDeployment.Target.DeploymentSettings.DisableAmLogForwarding, "for deployment", deployment.Name)
		changed = true
	}

	if newDeployment.Application.VCores != deployment.Application.VCores {
		fmt.Println("VCores changed from", deployment.Application.VCores, "to", newDeployment.Application.VCores, "for deployment", deployment.Name)
		changed = true
	}

	if newDeployment.Application.Ref.GroupID != deployment.Application.Ref.GroupID {
		fmt.Println("GroupID changed from", deployment.Application.Ref.GroupID, "to", newDeployment.Application.Ref.GroupID, "for deployment", deployment.Name)
		changed = true
	}

	if newDeployment.Application.Ref.ArtifactID != deployment.Application.Ref.ArtifactID {
		fmt.Println("ArtifactID changed from", deployment.Application.Ref.ArtifactID, "to", newDeployment.Application.Ref.ArtifactID, "for deployment", deployment.Name)
		changed = true
	}

	if newDeployment.Application.Ref.Packaging != deployment.Application.Ref.Packaging {
		fmt.Println("Packaging changed from", deployment.Application.Ref.Packaging, "to", newDeployment.Application.Ref.Packaging, "for deployment", deployment.Name)
		changed = true
	}

	if newDeployment.Application.Ref.Version != deployment.Application.Ref.Version {
		fmt.Println("Version changed from", deployment.Application.Ref.Version, "to", newDeployment.Application.Ref.Version, "for deployment", deployment.Name)
		changed = true
	}

	if schedulesHaveChanged(newDeployment.Application.Configuration.MuleAgentScheduleService.Schedulers, deployment.Application.Configuration.MuleAgentScheduleService.Schedulers) {
		fmt.Println("Schedulers Configuration changed for deployment", deployment.Name)
		changed = true
	}

	changed = changed || propertiesHaveChanged(
		deployment.Application.Configuration.MuleAgentApplicationPropertiesService.Properties,
		newDeployment.Application.Configuration.MuleAgentApplicationPropertiesService.Properties)

	return updatedDeployment, changed
}

// ingressUpdated returns true if the ingress settings have changed, false otherwise
func ingressUpdated(desiredHttpIngress, currentHttpIngress anypointclient.DeploymentHttpIngress) bool {
	// Check PathRewrite, LastMileSecurity, and ForwardSslSession
	if desiredHttpIngress.Inbound.PathRewrite != currentHttpIngress.Inbound.PathRewrite {
		log.Printf("PathRewrite changed from %q to %q", currentHttpIngress.Inbound.PathRewrite, desiredHttpIngress.Inbound.PathRewrite)
		return true
	}
	if desiredHttpIngress.Inbound.LastMileSecurity != currentHttpIngress.Inbound.LastMileSecurity {
		log.Printf("LastMileSecurity changed from %v to %v", currentHttpIngress.Inbound.LastMileSecurity, desiredHttpIngress.Inbound.LastMileSecurity)
		return true
	}
	if desiredHttpIngress.Inbound.ForwardSslSession != currentHttpIngress.Inbound.ForwardSslSession {
		log.Printf("ForwardSslSession changed from %v to %v", currentHttpIngress.Inbound.ForwardSslSession, desiredHttpIngress.Inbound.ForwardSslSession)
		return true
	}

	// Check InternalURL only if specified in desired config
	if desiredHttpIngress.Inbound.InternalURL != "" &&
		desiredHttpIngress.Inbound.InternalURL != currentHttpIngress.Inbound.InternalURL {
		log.Printf("InternalURL changed from %q to %q", currentHttpIngress.Inbound.InternalURL, desiredHttpIngress.Inbound.InternalURL)
		return true
	}

	// Compare PublicURL lists, filtering out auto-generated CloudHub URLs
	desiredURLs := filterNonCloudhubURLs(strings.Split(desiredHttpIngress.Inbound.PublicURL, ","))
	currentURLs := filterNonCloudhubURLs(strings.Split(currentHttpIngress.Inbound.PublicURL, ","))

	if !sliceEquals(desiredURLs, currentURLs) {
		log.Printf("PublicURL (filtered) changed from %v to %v", currentURLs, desiredURLs)
		return true
	}

	// If desired config specifies endpoints, compare only non-CloudHub endpoints
	if len(desiredHttpIngress.Inbound.Endpoints) > 0 {
		desiredEndpoints := filterNonCloudhubEndpoints(desiredHttpIngress.Inbound.Endpoints)
		currentEndpoints := filterNonCloudhubEndpoints(currentHttpIngress.Inbound.Endpoints)

		if len(desiredEndpoints) != len(currentEndpoints) {
			log.Printf("Endpoint count (filtered) changed from %d to %d", len(currentEndpoints), len(desiredEndpoints))
			return true
		}

		// Compare endpoints by matching URLs (not by index position)
		for _, desired := range desiredEndpoints {
			found := false
			for _, current := range currentEndpoints {
				if desired.URL == current.URL {
					found = true
					if !endpointEquals(desired, current) {
						log.Printf("Endpoint %q changed: PathRewrite %q->%q, Access %q->%q",
							desired.URL, current.PathRewrite, desired.PathRewrite, current.Access, desired.Access)
						return true
					}
					break
				}
			}
			if !found {
				log.Printf("Endpoint %q not found in current deployment", desired.URL)
				return true
			}
		}
	}

	return false
}

// runtimeVersionUpdated returns true if the version should be updated, false otherwise
//
// It determines if the runtime version should be updated by comparing the desiredVersion with the currentVersion
// The version naming convention for Mule runtimes version 4.5 and later in ClouHub, CloudHub 2.0 and Runtime Fabric is described on this page
// https://docs.mulesoft.com/release-notes/mule-runtime/updating-mule-4-versions#ensuring-mule-application-compatibility
//
// This function will always ignore the build and channel parts when comparing versions.
// If the desiredVersion starts with a "~" it will allow the patch-level of the current version to be higher than the desiredVersion.
func runtimeVersionUpdated(desiredVersion string, currentVersion string) bool {
	// Split the version strings by colon to remove the build and channel parts
	desiredParts := strings.SplitN(desiredVersion, ":", 2)
	currentParts := strings.SplitN(currentVersion, ":", 2)

	// Check if the desiredPatts[0] starts with a "~" if so we will allow patch-level to be higher than the desired version
	if desiredParts[0][0] == '~' {
		// Remove the "~" and split the remaining version by "."
		desiredParts[0] = desiredParts[0][1:]
		dParts := strings.SplitN(desiredParts[0], ".", 3)
		cParts := strings.SplitN(currentParts[0], ".", 3)
		// Check if the major and minor part of the version strings differs
		if dParts[0] != cParts[0] || dParts[1] != cParts[1] {
			return true
		}

		dPatchVersion, dErr := strconv.Atoi(dParts[2])
		cPatchVersion, cErr := strconv.Atoi(cParts[2])

		// If conversion fails, fall back to string comparison
		if dErr != nil || cErr != nil {
			return cParts[2] < dParts[2]
		}

		// Return true if current patch version is lower than desired patch version
		return cPatchVersion < dPatchVersion

	}

	// Compare the major, minor and patch part of the version strings
	return desiredParts[0] != currentParts[0]
}

func propertiesHaveChanged(oldProperties, newProperties map[string]string) bool {
	return propertiesMismatch(oldProperties, newProperties) || propertiesMismatch(newProperties, oldProperties)
}

func propertiesMismatch(src, dest map[string]string) bool {
	for property, value := range src {
		if dest[property] != value && !isSecret(value) {
			return true
		}
	}
	return false
}

func schedulesHaveChanged(slice1, slice2 []anypointclient.Schedule) bool {
	if len(slice1) != len(slice2) {
		return true
	}
	map1 := make(map[anypointclient.Schedule]struct{})
	map2 := make(map[anypointclient.Schedule]struct{})

	for _, obj := range slice1 {
		map1[obj] = struct{}{}
	}
	for _, obj := range slice2 {
		map2[obj] = struct{}{}
	}
	return !reflect.DeepEqual(map1, map2)
}

func isSecret(value string) bool {
	matching, _ := regexp.MatchString("^[*]+$", value)
	return matching
}

func GetAnypointClient() *anypointclient.AnypointClient {
	var err error
	var baseURL string

	if !viper.IsSet("base-url") {
		baseURL, err = anypointclient.ResolveBaseURLFromRegion(viper.GetString("region"))
		if err != nil {
			log.Fatalf("%e", err)
		}
	} else {
		baseURL = viper.GetString("base-url")
	}

	proxyURL := viper.GetString("proxy")

	switch viper.GetString("authType") {
	case "bearer":
		return anypointclient.NewAnypointClientWithToken(viper.GetString("bearer"), baseURL, proxyURL)
	case "user":
		return anypointclient.NewAnypointClientWithCredentials(viper.GetString("user"), viper.GetString("password"), baseURL, proxyURL)
	case "connectedapp":
		return anypointclient.NewAnypointClientWithConnectedApp(viper.GetString("client-id"), viper.GetString("client-secret"), baseURL, proxyURL)
	default:
		log.Fatalf("Unknown authentication method: %s", viper.GetString("authType"))
	}
	return nil
}

// sliceEquals compares two slices of comparable elements and returns true if they are equal, false otherwise.
func sliceEquals[T comparable](as, bs []T) bool {
	if len(as) != len(bs) {
		return false
	}
	if len(as) == 0 {
		return true
	}
	diff := make(map[T]int, len(as))
	for _, a := range as {
		diff[a]++
	}
	for _, b := range bs {
		current, ok := diff[b]
		if !ok {
			return false
		}
		if current == 1 {
			delete(diff, b)
			continue
		}
		diff[b] = current - 1
	}
	return len(diff) == 0
}

// endpointEquals compares two IngressEndpoint structs and returns true if they are equal
func endpointEquals(a, b anypointclient.IngressEndpoint) bool {
	return a.URL == b.URL &&
		a.PathRewrite == b.PathRewrite &&
		a.Access == b.Access
}

// isCloudHubURL returns true if the URL is an auto-generated CloudHub URL
func isCloudHubURL(url string) bool {
	return strings.Contains(url, ".cloudhub.io")
}

// filterNonCloudhubURLs returns only URLs that are not auto-generated CloudHub URLs
func filterNonCloudhubURLs(urls []string) []string {
	var filtered []string
	for _, url := range urls {
		if url != "" && !isCloudHubURL(url) {
			filtered = append(filtered, url)
		}
	}
	return filtered
}

// filterNonCloudhubEndpoints returns only endpoints that are not auto-generated CloudHub endpoints
func filterNonCloudhubEndpoints(endpoints []anypointclient.IngressEndpoint) []anypointclient.IngressEndpoint {
	var filtered []anypointclient.IngressEndpoint
	for _, endpoint := range endpoints {
		if !isCloudHubURL(endpoint.URL) {
			filtered = append(filtered, endpoint)
		}
	}
	return filtered
}
