# Anypoint CloudHub Deployment CLI

This is a command line tool to deploy new and update existing mule application running in Anypoint CloudHub using application artifacts stored in Exchange.

Currently it can only deploy applications that are published to Exchange. There is no support for uploading and deploying local Mule application artifacts.

It uses the Anypoint [Access Management API (Authentication)](https://anypoint.mulesoft.com/exchange/portals/anypoint-platform/f1e97bc6-315a-4490-82a7-23abe036327a.anypoint-platform/access-management-api/) and [CloudHub API](https://anypoint.mulesoft.com/exchange/portals/anypoint-platform/f1e97bc6-315a-4490-82a7-23abe036327a.anypoint-platform/cloudhub-api/) 

## Build instructions 

```shell
go get -d -v ./...
go build -o chdeploy
```

## Usage instructions

This tool uses json files to describe the desired deployments and detect if the desired deployment differs from the one actually running in Cloudhub. If so the running deployment will be updated.

### How to run (as user)
```shell
export ANYPOINT_USER=deployer
export ANYPOINT_PASSWORD=verysecretpassword
export ANYPOINT_ORGANIZATION_NAME=MasterOrg/SubBuissniessGroup
export ANYPOINT_ENVIRONMENT=Sandbox
export ANYPOINT_AUTH=user
./chdeploy -user $ANYPOINT_USER -password $ANYPOINT_PASSWORD -organization $ANYPOINT_ORGANIZATION_NAME -environment $ANYPOINT_ENVIRONMENT -authtype $ANYPOINT_AUTH  *.json 
```


### Deployment descriptors

The deployment descriptors are in JSON format and derived from the JSON payload handled by the Anypoint Cloudhub API. Below is an example.

```json
{
  "applicationSource" : {
      "artifactId" : "hellomuleworld",
      "groupId" : "deadbeef-abc1-3743-aX61-71337e847d4",
      "organizationId" : "deadbeef-abc1-3743-aX61-71337e847d4",
      "source" : "EXCHANGE",
      "version" : "1.0.0"
   },
   "applicationInfo" : {
      "domain" : "helloworld-ullpo-prod",
      "logLevels" : [],
      "loggingCustomLog4JEnabled" : false,
      "loggingNgEnabled" : true,
      "monitoringAutoRestart" : true,
      "monitoringEnabled" : true,
      "muleVersion" : {
         "version" : "4.4.0"
      },
      "objectStoreV1" : false,
      "persistentQueues" : false,
      "persistentQueuesEncrypted" : false,
      "properties" : {
         "anypoint.platform.config.analytics.agent.enabled" : "true",
         "anypoint.platform.visualizer.layer" : "System",
         "servicename": "production environment",
         "backend.endpoint": "https://httpbin.org/anything"
      },
      "staticIPsEnabled" : false,
      "trackingSettings" : {
         "trackingLevel" : "DISABLED"
      },
      "workers" : {
         "amount" : 1,
         "type" : {
            "name" : "Micro"
         }
      }
   },
   
   "autoStart" : true
}
```

## Contribution Guidelines

All contributions are welcome. Feel free to open a pull request.
