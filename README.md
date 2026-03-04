# Anypoint CloudHub Deployment CLI

This is a command line tool to deploy new and update existing resources in Anypoint Platform.

	The tool supports
	* Mule application running in Anypoint CloudHub 2.0 using application artifacts stored in Exchange
	* API instance policies for APIs managed in Anypoint API manager
	* Anypoint MQ destinations (queues, exchanges, and bindings)

It uses the Anypoint [Access Management API (Authentication)](https://anypoint.mulesoft.com/exchange/portals/anypoint-platform/f1e97bc6-315a-4490-82a7-23abe036327a.anypoint-platform/access-management-api/) and [CloudHub API](https://anypoint.mulesoft.com/exchange/portals/anypoint-platform/f1e97bc6-315a-4490-82a7-23abe036327a.anypoint-platform/cloudhub-api/) 

## Run tests 

To run the complete test suite use the build in go test subcommand

```shell
go test -v ./...
```

## Build instructions 

```shell
go get -d -v ./...
go build -o chdeploy
```

#### Build instructions for QA and PROD deployer

```shell
GOOS=linux GOARCH=amd64 go build -o anypointchdeployer
```
Ensures compatibility with QA and PROD pipeline servers by building the anypointchdeployer executable for Linux (AMD64). 

## Usage instructions

This tool uses json files to describe the desired deployments and detect if the desired deployment differs from the one actually running in Cloudhub. If so the running deployment will be updated.

### How to run (as user)
```shell
./chdeploy -u <username> -p <password> -o <organizationname> -e <environment> -a user  *.json
```

### Dry run mode

Use the `--dry-run` flag to see what changes would be made without actually applying them:

```shell
./chdeploy -o <organizationname> -e <environment> --dry-run *.json
```

### MQ destinations

For MQ destinations, specify the MQ region with `--mq-region`:

```shell
./chdeploy -o <organizationname> -e <environment> -m eu-central-1 mq-destinations.json
```

### Deployment descriptors

#### Application Deployment descriptors

The deployment descriptors are in JSON format and derived from the JSON payload handled by the Anypoint Cloudhub API. Below is an example.


```json
{
  "kind": "Application",
  "version": "v1",
  "spec": {
    "name": "tjenna",
    "target": {
      "provider": "MC",
      "targetId": "23d6f918-e088-4ab7-ad56-d9a62328bfec",
      "deploymentSettings": {
        "clustered": false,
        "enforceDeployingReplicasAcrossNodes": true,
        "http": {
          "inbound": {
            "publicUrl": "https://tesing2-fsuil6.y5muwb.deu-c1.cloudhub.io/",
            "pathRewrite": null,
            "lastMileSecurity": false,
            "forwardSslSession": false
          }
        },
        "jvm": {},
        "outbound": {},
        "runtime": {
          "version": "4.6.10",
          "releaseChannel": "LTS",
          "java": "17"
        },
        "updateStrategy": "rolling",
        "disableAmLogForwarding": false,
        "persistentObjectStore": false,
        "generateDefaultPublicUrl": true
      },
      "replicas": 1
    },
    "application": {
      "ref": {
        "groupId": "78368f4a-979c-4b2f-9445-01d8e2f35426",
        "artifactId": "debug-tools-mule4-1.0.0-SNAPSHOT-mule-application",
        "version": "1.0.0",
        "packaging": "jar"
      },
      "assets": [],
      "desiredState": "STARTED",
      "configuration": {
        "mule.agent.application.properties.service": {
          "applicationName": "testing2",
          "properties": {
            "sometest": "change"
          },
          "secureProperties": {}
        },
        "mule.agent.logging.service": {
          "scopeLoggingConfigurations": []
        },
        "mule.agent.scheduling.service": {
          "schedulers": []
        }
      },
      "integrations": {
        "services": {
          "objectStoreV2": {
            "enabled": false
          }
        }
      },
      "vCores": 0.1
    }
  }
}
```


##### Mule runtime version tilde ranges

The `target.runtime.version` field supports "tilde ranges" for the runtime version. This means that by prefixing the requested 
version with a tilde (`~`) Anypoint CH Deployer will allow the running version to be any patchlevel higher than the one specified.

Changes in major or minor versions will result in a update (either upgrade or downgrade).

<table>
<tr>
<th>version defined in json file</th><th>version running in cloudhub</th><th>Will update</th>
</tr>
<tr>
<td>4.6.12</td><td>4.6.12:1</td><td>No</td>
</tr>
<tr>
<td>4.6.12</td><td>4.6.14:3</td><td>Yes will downgrade</td>
</tr>
<tr>
<td>~4.6.12</td><td>4.6.14:3</td><td>No</td>
</tr>
<tr>
<td>~4.6.12</td><td>4.6.10:1</td><td>Yes will upgrade</td>
</tr>
<tr>
<td>~4.6.12</td><td>4.4.0:20250217-2</td><td>Yes will upgrade</td>
</tr>
<tr>
<td>~4.6.12</td><td>4.9.2:6</td><td>Yes will downgrade</td>
</tr>
</table>


```json
{
  "kind": "Application",
  "version": "v1",
  "spec": {
    "name": "app-name",
    "target": {
      "provider": "MC",
      "targetId": "9c237504-4796-401b-9dbd-689f1857edcf",
      "deploymentSettings": {
        "clustered": false,
        "enforceDeployingReplicasAcrossNodes": false,
        "http": {
          "inbound": {
            "publicUrl": "",
            "pathRewrite": "",
            "lastMileSecurity": false,
            "forwardSslSession": false,
            "internalUrl": "",
            "uniqueId": ""
          }
        },
        "jvm": {
          "args": ""
        },
        "runtime": {
          "version": "",
          "releaseChannel": "",
          "java": ""
        },
        "updateStrategy": "",
        "resources": {
          "cpu": {
            "limit": "",
            "reserved": ""
          },
          "memory": {
            "limit": "",
            "reserved": ""
          },
          "storage": {
            "limit": "",
            "reserved": ""
          }
        },
        "lastMileSecurity": false,
        "disableAmLogForwarding": false,
        "persistentObjectStore": false,
        "anypointMonitoringScope": "",
        "sidecars": {
          "anypoint-monitoring": {
            "image": "",
            "resources": {
              "cpu": {
                "limit": "",
                "reserved": ""
              },
              "memory": {
                "limit": "",
                "reserved": ""
              },
              "storage": {
                "limit": "",
                "reserved": ""
              }
            }
          }
        },
        "forwardSslSession": false,
        "disableExternalLogForwarding": false,
        "generateDefaultPublicUrl": false,
        "cpuMax": "",
        "cpuReserved": "",
        "memoryMax": "",
        "memoryReserved": "",
        "replicationFactor": 0,
        "environmentVariables": {},
        "clusteringEnabled": false,
        "publicUrl": "",
        "jvmProperties": ""
      },
      "replicas": 1
    },
    "application": {
      "desiredState": "STARTED",
      "ref": {
        "classifier": "",
        "type": "",
        "groupId": "",
        "artifactId": "",
        "version": "",
        "packaging": ""
      },
      "configuration": {
        "mule.agent.application.properties.service": {
          "applicationName": "",
          "properties": {},
          "secureProperties": {},
          "secureproperties": {}
        },
        "mule.agent.logging.service": {
          "artifactName": "",
          "scopeLoggingConfigurations": [
            {
              "scope": "",
              "logLevel": ""
            }
          ]
        },
        "mule.agent.scheduling.service": {
          "applicationName": "",
          "schedulers": [
            {
              "name": "",
              "type": "",
              "flowName": "",
              "enabled": false,
              "timeUnit": "NANOSECONDS",
              "frequency": "",
              "startDelay": "",
              "expression": "",
              "timeZone": ""
            }
          ]
        }
      },
      "resourceAssets": {},
      "vCores": 0,
      "integrations": {
        "services": {
          "objectStoreV2": {
            "enabled": false
          }
        }
      },
      "objectStoreV2Enabled": false
    },
    "desiredVersion": "1.0.0"
  }
}
```

#### API Policy Deployment descriptors

The deployment descriptors are in JSON format and derived from the JSON payload handled by the Anypoint ApiManafer API. Below is an example.


```json
{
  "kind": "ApiPolicies",
  "version": "v1",
  "spec": {
      "apiInstanceId": "2582478",
      "policy": {
          "pointcutData": null,
          "groupId": "e0b4a150-f59b-46d4-ad25-5d98f9deb24a",
          "assetId": "ip-allowlist",
          "assetVersion": "1.1.1",
          "configurationData": {
              "ipExpression": "#[attributes.headers['x-forwarded-for']]",
              "ips": [
                  "127.0.0.1",
                  "192.0.2.1",
                  "198.51.100.2",
                  "203.0.113.3"
              ]
          }
      }
  }
}
```

#### MQ Destinations Deployment descriptors

MQ destinations support queues, exchanges, and exchange bindings with routing rules.

```json
{
  "version": "v1",
  "kind": "MqDestinations",
  "spec": {
    "queues": [
      {
        "queueId": "my-queue-dlq",
        "type": "queue",
        "fifo": false,
        "encrypted": true,
        "defaultTtl": 604800000,
        "defaultLockTtl": 120000,
        "defaultDeliveryDelay": 0
      },
      {
        "queueId": "my-queue",
        "type": "queue",
        "fifo": false,
        "encrypted": true,
        "maxDeliveries": 2,
        "deadLetterQueueId": "my-queue-dlq",
        "defaultTtl": 604800000,
        "defaultLockTtl": 120000,
        "defaultDeliveryDelay": 0
      }
    ],
    "exchanges": [
      {
        "exchangeId": "my-exchange",
        "fifo": false,
        "encrypted": true,
        "bindings": [
          {
            "queueId": "my-queue",
            "fifo": false,
            "routingRules": [
              {
                "propertyName": "type",
                "propertyType": "STRING",
                "matcherType": "EQ",
                "value": "order"
              }
            ]
          }
        ]
      }
    ]
  }
}
```

**Queue validation ranges:**
- `maxDeliveries`: 1-1000
- `defaultTtl`: 60000-1209600000 ms (1 min to 14 days)
- `defaultLockTtl`: 0-43200000 ms (0 to 12 hours)
- `defaultDeliveryDelay`: 0 or 1000-900000 ms (0 or 1 sec to 15 min)

## Contribution Guidelines

All contributions are welcome. Feel free to open a pull request.
