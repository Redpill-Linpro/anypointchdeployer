package appconf

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/Redpill-Linpro/anypointchdeployer/pkg/anypointclient"
	"github.com/spf13/viper"
)

func ApplicationHasChanged(application anypointclient.CloudhubApplicationResponse, newApplication anypointclient.CloudhubApplicationRequest) bool {

	newConf := newApplication.ApplicationInfo
	oldConf := application

	oldProperties := make(map[string]string)
	applicationData, _ := json.Marshal(application.Properties)
	json.Unmarshal(applicationData, &oldProperties)

	newProperties := make(map[string]string)
	newApplicationData, _ := json.Marshal(newApplication.ApplicationInfo.Properties)
	json.Unmarshal(newApplicationData, &newProperties)

	changed := (newConf.Domain == oldConf.Domain) &&
		(newConf.LoggingCustomLog4JEnabled == oldConf.LoggingCustomLog4JEnabled) &&
		(newConf.LoggingNgEnabled == oldConf.LoggingNgEnabled) &&
		(newConf.MonitoringAutoRestart == oldConf.MonitoringAutoRestart) &&
		(newConf.MonitoringEnabled == oldConf.MonitoringEnabled) &&
		(newConf.PersistentQueues == oldConf.PersistentQueues) &&
		(newConf.PersistentQueuesEncrypted == oldConf.PersistentQueuesEncrypted) &&
		(newConf.StaticIPsEnabled == oldConf.StaticIPsEnabled) &&
		(newConf.Workers.Amount == oldConf.Workers.Amount) &&
		(newConf.Workers.Type.Name == oldConf.Workers.Type.Name)

	changed = changed && propertiesHasChanged(oldProperties, newProperties)

	return true
}

func propertiesHasChanged(oldProperties map[string]string, newProperties map[string]string) bool {

	for property, value := range oldProperties {
		if newProperties[property] != value {
			//  TO-DO Temporary hack to skip secret properties
			matching, _ := regexp.MatchString("^[*]+$", newProperties[property])
			if !matching {
				return true
			}
		}
	}
	return false
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

	switch viper.GetString("authType") {
	case "bearer":
		return anypointclient.NewAnypointClientWithToken(viper.GetString("bearer"), baseURL)
	case "user":
		return anypointclient.NewAnypointClientWithCredentials(viper.GetString("user"), viper.GetString("password"), baseURL)
	case "connectedapp":
		return anypointclient.NewAnypointClientWithConnectedApp(viper.GetString("client-id"), viper.GetString("client-secret"), baseURL)
	default:
		log.Fatalf("Unknown authentication method: %s", viper.GetString("authType"))
	}
	return nil
}

func AddApplicationGav(app anypointclient.CloudhubApplicationRequest) {
	groupId := app.ApplicationSource.GroupID
	artifactId := app.ApplicationSource.ArtifactID
	version := app.ApplicationSource.Version

	app.ApplicationInfo.Properties["chdeployer.application.gav"] = fmt.Sprintf("%s:%s:%s", groupId, artifactId, version)
}
