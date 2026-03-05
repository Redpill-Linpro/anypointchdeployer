package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"sync"

	"github.com/Redpill-Linpro/anypointchdeployer/internal/appconf"
	"github.com/Redpill-Linpro/anypointchdeployer/internal/flagvalidator"
	"github.com/Redpill-Linpro/anypointchdeployer/internal/resources"
	"github.com/Redpill-Linpro/anypointchdeployer/pkg/anypointclient"
	"github.com/TwiN/go-color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "anypointchdeployer",
	Short: "CLI tool to deploy and update resources in Anypoint Platform",
	Long: `This is a command line tool to deploy new and update existing resources in Anypoint Platform. 
	
	The tool supports 
	* Mule application running in Anypoint CloudHub 2.0 using application artifacts stored in Exchange
	* API instance policies for APIs managed in Anypoint API manager
	`,
	Example:   "./chdeploy -u <username> -p <password> -o <organizationname> -e <environment> *.json",
	ValidArgs: []string{"*.json"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := flagvalidator.ValidateFlags(); err != nil {
			log.Fatalf("%+v\n", err)
		}
		client := appconf.GetAnypointClient()

		err := client.Login()
		if err != nil {
			log.Fatalf("Fail to login to anypoint platform %+v\n", err)
		}
		organization, err := client.ResolveOrganization(viper.GetString("organization"))
		if err != nil {
			log.Fatalf("failed to get organization %+v", err)
		}

		environment, err := client.ResolveEnvironment(organization, viper.GetString("environment"))
		if err != nil {
			log.Fatalf("failed to get environment %+v", err)
		}
		var privateSpace anypointclient.PrivateSpace = anypointclient.PrivateSpace{}
		if viper.GetString("private-space") != "" {
			privateSpace, err = client.ResolvePrivateSpace(organization, viper.GetString("private-space"))
			if err != nil {
				log.Fatalf("failed to get private space %+v", err)
			}
		}
		deployConfig(client, args, organization, environment, privateSpace)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("region", "r", "US", "region for Anypoint. Use US for US control plane and EU for EU control plane")
	rootCmd.Flags().StringP("base-url", "l", "", "base url for Anypoint platform")
	rootCmd.Flags().StringP("proxy", "x", "", "HTTP proxy URL (e.g., http://proxy:8080)")
	rootCmd.Flags().StringP("authtype", "a", "connectedapp", "authentication method towards Anypoint Platform")
	rootCmd.Flags().StringP("bearer", "b", "", "authentication bearer token used to authenticate with Anypoint")
	rootCmd.Flags().StringP("user", "u", "", "user to use to login to Anypoint if token is not provided")
	rootCmd.Flags().StringP("password", "p", "", "password for the Anypoint user")
	rootCmd.Flags().StringP("client-id", "i", "", "client id for the Anypoint connected app")
	rootCmd.Flags().StringP("client-secret", "s", "", "client secret for the Anypoint connected app")
	rootCmd.Flags().StringP("organization", "o", "", "organization within Anypoint Platform")
	rootCmd.Flags().StringP("environment", "e", "", "environment within Anypoint Platform")
	rootCmd.Flags().StringP("private-space", "v", "", "private space within Anypint Platform")
	rootCmd.Flags().BoolP("force-update", "f", false, "force update even if no changes are detected")
	rootCmd.Flags().Bool("dry-run", false, "show what would be done without making any changes")
	rootCmd.Flags().IntP("concurrent-deployments", "c", 1, "max number of concurrent deploys")
	rootCmd.Flags().StringP("mq-region", "m", "", "MQ region for Anypoint MQ destinations (e.g., eu-west-1, us-east-1)")
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(f.Name, f)
	})
	flagvalidator.AddFlagSetValidator("region", []any{"US", "EU"})
	flagvalidator.AddFlagSetValidator("authType", []any{"bearer", "user", "connectedapp"})
	flagvalidator.AddFlagSetValidator("concurrent-deployments", []any{1, 2, 3, 4, 5})
}

func deployConfig(client *anypointclient.AnypointClient, files []string, organization anypointclient.Organization, environment anypointclient.Environment, privateSpace anypointclient.PrivateSpace) {

	var wg sync.WaitGroup
	guard := make(chan struct{}, viper.GetInt("concurrent-deployments"))
	faults := make(chan error, len(files))

	defer func() {
		close(guard)
	}()

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			guard <- struct{}{}
			defer func() {
				wg.Done()
				<-guard
			}()

			log.Printf("Reading file: %s", file)

			fileData, err := os.ReadFile(file)
			if err != nil {
				faults <- fmt.Errorf("failed to open file: %s. Error: %v", file, err)
				return
			}
			expandedData := os.ExpandEnv(string(fileData))

			resource, err := unmarshalResource([]byte(expandedData))
			if err != nil {
				faults <- fmt.Errorf("failed to decode %v %+v", flag.Arg(0), err)
				return
			}
			switch r := resource.(type) {
			case resources.ApplicationV1:
				err = deployApplication(r.Spec, client, organization, environment, privateSpace)
				if err != nil {
					faults <- err
					return
				}
			case resources.ApiPoliciesV1:
				err = deployApiPolicy(r, client, organization, environment)
				if err != nil {
					faults <- err
					return
				}
			case resources.MqDestinationsV1:
				err = deployMqDestinations(r, client, organization, environment)
				if err != nil {
					faults <- err
					return
				}
			}

		}(file)
	}
	wg.Wait()
	close(faults)
	if len(faults) > 0 {
		for fault := range faults {
			log.Println(color.Colorize(color.Red, fmt.Sprintf("%+v\n", fault)))
		}
		os.Exit(10)
	}
	log.Println(color.Colorize(color.Green, "All deployments handled successfully!\n"))
}

func unmarshalResource(data []byte) (any, error) {
	var vr resources.BaseResource
	if err := json.Unmarshal(data, &vr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal version: %w", err)
	}

	switch vr.Kind {
	case "":
		return nil, fmt.Errorf("missing required field 'kind' in resource definition")
	case "Application":
		switch vr.Version {
		case "v1":
			var r resources.ApplicationV1
			if err := json.Unmarshal(data, &r); err != nil {
				return nil, fmt.Errorf("failed to unmarshal ApplicationV1: %w", err)
			}
			return r, nil
		default:
			return nil, fmt.Errorf("unknown application version: %s", vr.Version)
		}

	case "ApiPolicies":
		switch vr.Version {
		case "v1":
			var r resources.ApiPoliciesV1
			if err := json.Unmarshal(data, &r); err != nil {
				return nil, fmt.Errorf("failed to unmarshal ApiPoliciesV1: %w", err)
			}
			return r, nil
		default:
			return nil, fmt.Errorf("unknown api policies version: %s", vr.Version)
		}

	case "MqDestinations":
		switch vr.Version {
		case "v1":
			var r resources.MqDestinationsV1
			if err := json.Unmarshal(data, &r); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MqDestinationsV1: %w", err)
			}
			return r, nil
		default:
			return nil, fmt.Errorf("unknown MQ destinations version: %s", vr.Version)
		}
	default:
		return nil, fmt.Errorf("unknown kind: %s", vr.Kind)
	}
}

func deployApplication(newDeployment anypointclient.CloudhubDeploymentReq, client *anypointclient.AnypointClient, organization anypointclient.Organization, environment anypointclient.Environment, privateSpace anypointclient.PrivateSpace) error {
	// Update the deployment to match latest schema version
	updatedDeployment, err := appconf.UpdateDeploymentToLatestSchema(newDeployment)
	if err != nil {
		return fmt.Errorf("failed to update deployment schema: %v", err)
	}

	client.UpdateScheduleNames(updatedDeployment.Application.Configuration.MuleAgentScheduleService.Schedulers)

	log.Println(color.Colorize(color.Green, fmt.Sprintf("Will deploy version [%s]", updatedDeployment.Application.Ref.Version)))

	deployment, err := client.GetDeployment(environment, updatedDeployment.Name)
	if err != nil {
		return fmt.Errorf("failed to get deployment %+v", err)
	}

	dryRun := viper.GetBool("dry-run")

	if deployment.Name == "" {
		if dryRun {
			log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would CREATE deployment: [%s]", updatedDeployment.Name)))
			return nil
		}
		deployment, err := client.CreateDeployment(environment, privateSpace, updatedDeployment)
		if err != nil {
			return fmt.Errorf("failed to create deployment %+v", err)
		}
		log.Println(color.Colorize(color.Green, fmt.Sprintf("Deployment: [%s] successfully created", updatedDeployment.Name)))

		if len(updatedDeployment.Application.Configuration.MuleAgentScheduleService.Schedulers) > 0 {
			log.Println(color.Colorize(color.Green, "Schedulers Configuration: Check that FlowName and Type match source code"))
			err = client.SchedulesDiffFromSourceCode(environment, updatedDeployment, deployment.ID)
			if err != nil {
				return fmt.Errorf("%s\ncause: %+v", updatedDeployment.Name, err)
			}
			log.Println(color.Colorize(color.Green, "Scheduler Configuration: Configurations match successfully"))
		}
		return nil
	}

	updatedDeployment, shouldUpdate := appconf.PrepareDeploymentAndCheckChanges(updatedDeployment, deployment)
	if shouldUpdate || viper.GetBool("force-update") {
		if dryRun {
			log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would UPDATE deployment: [%s]", updatedDeployment.Name)))
			return nil
		}
		err := client.UpdateDeployment(environment, privateSpace, updatedDeployment, deployment.ID)
		if err != nil {
			return fmt.Errorf("failed to update application: %s\ncause: %+v", updatedDeployment.Name, err)
		}
		if viper.GetBool("force-update") {
			log.Println(color.Colorize(color.Green, fmt.Sprintf("Deployment: [%s] successfully forced updated", deployment.Name)))
		} else {
			log.Println(color.Colorize(color.Green, fmt.Sprintf("Deployment: [%s] successfully updated", deployment.Name)))
		}

		if len(updatedDeployment.Application.Configuration.MuleAgentScheduleService.Schedulers) > 0 {
			log.Println(color.Colorize(color.Green, "Schedulers Configuration: Check that FlowName and Type match source code"))
			err = client.SchedulesDiffFromSourceCode(environment, updatedDeployment, deployment.ID)
			if err != nil {
				return fmt.Errorf("%s\ncause: %+v", updatedDeployment.Name, err)
			}
			log.Println(color.Colorize(color.Green, "Scheduler Configuration: Configurations match successfully"))
		}
		return nil
	}
	log.Println(color.Colorize(color.Blue, fmt.Sprintf("Deployment: [%s] already deployed with correct configuration", deployment.Name)))
	return nil
}

func deployApiPolicy(apipolicies resources.ApiPoliciesV1, client *anypointclient.AnypointClient, organization anypointclient.Organization, environment anypointclient.Environment) error {
	log.Printf("Deploying API policies on API instance %s\n", apipolicies.Spec.ApiInstanceID)
	dryRun := viper.GetBool("dry-run")

	// Get API instance ID from the spec
	apiInstanceID, err := strconv.Atoi(apipolicies.Spec.ApiInstanceID)
	if err != nil {
		return fmt.Errorf("invalid API instance ID: %s, error: %v", apipolicies.Spec.ApiInstanceID, err)
	}

	// Get existing policies for this API instance
	existingPolicies, err := client.GetApiInstancePolicies(organization.ID, environment.ID, apiInstanceID)
	if err != nil {
		return fmt.Errorf("failed to get API instance policies: %v", err)
	}

	for _, apipolicy := range apipolicies.Spec.Policies {

		var matchingPolicy *anypointclient.ApiPolicyResponse
		for i, policy := range *existingPolicies {
			// Check if policy matches the Group ID, AssetID, and PointcutData in the spec
			if policy.Template.GroupID == apipolicy.GroupID &&
				policy.Template.AssetID == apipolicy.AssetID &&
				reflect.DeepEqual(policy.PointcutData, apipolicy.PointcutData) {
				matchingPolicy = &(*existingPolicies)[i]
				break
			}
		}

		// No policy with the same Group ID, Asset ID, and pointcut is found, create a new one
		if matchingPolicy == nil {
			log.Println(color.Colorize(color.Yellow, fmt.Sprintf("Policy with matching template %s:%s:%s and pointcut not found for API instance %d", apipolicy.GroupID, apipolicy.AssetID, apipolicy.AssetVersion, apiInstanceID)))
			if dryRun {
				log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would CREATE API Policy %s:%s:%s for instance %d", apipolicy.GroupID, apipolicy.AssetID, apipolicy.AssetVersion, apiInstanceID)))
				continue
			}
			err = client.CreateApiInstancePolicies(
				organization.ID,
				environment.ID,
				apiInstanceID,
				apipolicy,
			)

			if err != nil {
				return fmt.Errorf("failed to create API policy for API instance %d: %v", apiInstanceID, err)
			}

			log.Println(color.Colorize(color.Green, fmt.Sprintf("API Policy %s:%s:%s for instance %d successfully created", apipolicy.GroupID, apipolicy.AssetID, apipolicy.AssetVersion, apiInstanceID)))
			continue
		}

		// There is an existing policy with the same Group ID, Asset ID, and pointcut, update it if necessary
		// If there is no asserVersion in the file use the current version
		if apipolicy.AssetVersion == "" {
			apipolicy.AssetVersion = matchingPolicy.Template.AssetVersion
		}
		// Check if version or configuration data has changed
		configChanged := false
		if matchingPolicy.Template.AssetVersion != apipolicy.AssetVersion || !reflect.DeepEqual(matchingPolicy.Configuration, apipolicy.ConfigurationData) {
			configChanged = true
		}

		// Update policy if configuration has changed
		if configChanged || viper.GetBool("force-update") {
			if dryRun {
				log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would UPDATE API Policy %s:%s:%s for instance %d", apipolicy.GroupID, apipolicy.AssetID, apipolicy.AssetVersion, apiInstanceID)))
				continue
			}
			// Update the policy with new configuration
			err = client.UpdateApiInstancePolicies(
				organization.ID,
				environment.ID,
				apiInstanceID,
				matchingPolicy.PolicyID,
				apipolicy,
			)

			if err != nil {
				return fmt.Errorf("failed to update API policy for API instance %d: %v", apiInstanceID, err)

			}

			if viper.GetBool("force-update") {
				log.Println(color.Colorize(color.Green, fmt.Sprintf("API Policy %s:%s:%s  for instance %d successfully forced updated", apipolicy.GroupID, apipolicy.AssetID, apipolicy.AssetVersion, apiInstanceID)))
			} else {
				log.Println(color.Colorize(color.Green, fmt.Sprintf("API Policy %s:%s:%s for instance %d successfully updated", apipolicy.GroupID, apipolicy.AssetID, apipolicy.AssetVersion, apiInstanceID)))
			}
		} else {
			log.Println(color.Colorize(color.Blue, fmt.Sprintf("API Policy %s:%s:%s for instance %d already configured correctly", apipolicy.GroupID, apipolicy.AssetID, apipolicy.AssetVersion, apiInstanceID)))
		}
	}

	return nil
}

func deployMqDestinations(mqDestinations resources.MqDestinationsV1, client *anypointclient.AnypointClient, organization anypointclient.Organization, environment anypointclient.Environment) error {
	mqRegion := viper.GetString("mq-region")
	if mqRegion == "" {
		return fmt.Errorf("--mq-region flag is required for MqDestinations resources")
	}
	dryRun := viper.GetBool("dry-run")

	log.Printf("Deploying MQ destinations to region %s\n", mqRegion)

	// Sort queues: DLQs first (queues without deadLetterQueueId), then queues that reference DLQs
	sortedQueues := make([]anypointclient.MqQueue, 0, len(mqDestinations.Spec.Queues))
	var queuesWithDLQ []anypointclient.MqQueue
	for _, queue := range mqDestinations.Spec.Queues {
		if queue.DeadLetterQueueID == "" {
			sortedQueues = append(sortedQueues, queue)
		} else {
			queuesWithDLQ = append(queuesWithDLQ, queue)
		}
	}
	sortedQueues = append(sortedQueues, queuesWithDLQ...)

	// Deploy queues
	for _, queue := range sortedQueues {
		if err := queue.Validate(); err != nil {
			return err
		}

		existingQueue, err := client.GetMqQueue(organization.ID, environment.ID, mqRegion, queue.QueueID)
		if err != nil {
			return fmt.Errorf("failed to get queue %s: %v", queue.QueueID, err)
		}

		if existingQueue == nil {
			if dryRun {
				log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would CREATE queue: [%s]", queue.QueueID)))
			} else {
				log.Printf("Creating queue: %s\n", queue.QueueID)
				err = client.CreateMqQueue(organization.ID, environment.ID, mqRegion, queue)
				if err != nil {
					return fmt.Errorf("failed to create queue %s: %v", queue.QueueID, err)
				}
				log.Println(color.Colorize(color.Green, fmt.Sprintf("Queue [%s] successfully created", queue.QueueID)))
			}
		} else if queueNeedsUpdate(queue, *existingQueue) || viper.GetBool("force-update") {
			if dryRun {
				log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would UPDATE queue: [%s]", queue.QueueID)))
			} else {
				log.Printf("Updating queue: %s\n", queue.QueueID)
				err = client.UpdateMqQueue(organization.ID, environment.ID, mqRegion, queue)
				if err != nil {
					return fmt.Errorf("failed to update queue %s: %v", queue.QueueID, err)
				}
				log.Println(color.Colorize(color.Green, fmt.Sprintf("Queue [%s] successfully updated", queue.QueueID)))
			}
		} else {
			log.Println(color.Colorize(color.Blue, fmt.Sprintf("Queue [%s] already configured correctly", queue.QueueID)))
		}
	}

	// Deploy exchanges and their bindings
	for _, exchange := range mqDestinations.Spec.Exchanges {
		existingExchange, err := client.GetMqExchange(organization.ID, environment.ID, mqRegion, exchange.ExchangeID)
		if err != nil {
			return fmt.Errorf("failed to get exchange %s: %v", exchange.ExchangeID, err)
		}

		if existingExchange == nil {
			if dryRun {
				log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would CREATE exchange: [%s]", exchange.ExchangeID)))
			} else {
				log.Printf("Creating exchange: %s\n", exchange.ExchangeID)
				err = client.CreateMqExchange(organization.ID, environment.ID, mqRegion, exchange.MqExchange)
				if err != nil {
					return fmt.Errorf("failed to create exchange %s: %v", exchange.ExchangeID, err)
				}
				log.Println(color.Colorize(color.Green, fmt.Sprintf("Exchange [%s] successfully created", exchange.ExchangeID)))
			}
		} else if exchangeNeedsUpdate(exchange.MqExchange, *existingExchange) || viper.GetBool("force-update") {
			if dryRun {
				log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would UPDATE exchange: [%s]", exchange.ExchangeID)))
			} else {
				log.Printf("Updating exchange: %s\n", exchange.ExchangeID)
				err = client.CreateMqExchange(organization.ID, environment.ID, mqRegion, exchange.MqExchange)
				if err != nil {
					return fmt.Errorf("failed to update exchange %s: %v", exchange.ExchangeID, err)
				}
				log.Println(color.Colorize(color.Green, fmt.Sprintf("Exchange [%s] successfully updated", exchange.ExchangeID)))
			}
		} else {
			log.Println(color.Colorize(color.Blue, fmt.Sprintf("Exchange [%s] already configured correctly", exchange.ExchangeID)))
		}

		// Handle bindings for this exchange
		err = syncExchangeBindings(client, organization.ID, environment.ID, mqRegion, exchange.ExchangeID, exchange.Bindings, dryRun)
		if err != nil {
			return fmt.Errorf("failed to sync bindings for exchange %s: %v", exchange.ExchangeID, err)
		}
	}

	log.Println(color.Colorize(color.Green, "MQ destinations deployment completed successfully"))
	return nil
}

func queueNeedsUpdate(desired anypointclient.MqQueue, current anypointclient.MqDestination) bool {
	return desired.Fifo != current.Fifo ||
		desired.IsEncrypted() != current.Encrypted ||
		desired.MaxDeliveries != current.MaxDeliveries ||
		desired.DeadLetterQueueID != current.DeadLetterQueueID ||
		desired.IsFallback != current.IsFallback ||
		desired.DefaultTtl != current.DefaultTtl ||
		desired.DefaultLockTtl != current.DefaultLockTtl ||
		desired.DefaultDeliveryDelay != current.DefaultDeliveryDelay
}

func exchangeNeedsUpdate(desired anypointclient.MqExchange, current anypointclient.MqDestination) bool {
	return desired.Fifo != current.Fifo || desired.IsEncrypted() != current.Encrypted
}

func syncExchangeBindings(client *anypointclient.AnypointClient, orgID, envID, region, exchangeID string, desiredBindings []anypointclient.MqBinding, dryRun bool) error {
	existingBindings, err := client.GetMqExchangeBindings(orgID, envID, region, exchangeID)
	if err != nil {
		return fmt.Errorf("failed to get existing bindings: %v", err)
	}

	// Create a map of existing bindings by queueID for quick lookup
	existingBindingsMap := make(map[string]anypointclient.MqBinding)
	if existingBindings != nil {
		for _, binding := range existingBindings {
			existingBindingsMap[binding.QueueID] = binding
		}
	}

	// Create a map of desired bindings by queueID
	desiredBindingsMap := make(map[string]anypointclient.MqBinding)
	for _, binding := range desiredBindings {
		desiredBindingsMap[binding.QueueID] = binding
	}

	// Create or update bindings
	for _, desiredBinding := range desiredBindings {
		existingBinding, exists := existingBindingsMap[desiredBinding.QueueID]
		if !exists {
			if dryRun {
				log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would CREATE binding: [%s -> %s]", exchangeID, desiredBinding.QueueID)))
				if len(desiredBinding.RoutingRules) > 0 {
					log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would SET routing rules for: [%s -> %s]", exchangeID, desiredBinding.QueueID)))
				}
			} else {
				// Create binding first (no body)
				log.Printf("Creating binding: %s -> %s\n", exchangeID, desiredBinding.QueueID)
				err = client.CreateMqBinding(orgID, envID, region, exchangeID, desiredBinding.QueueID)
				if err != nil {
					return fmt.Errorf("failed to create binding for queue %s: %v", desiredBinding.QueueID, err)
				}
				log.Println(color.Colorize(color.Green, fmt.Sprintf("Binding [%s -> %s] successfully created", exchangeID, desiredBinding.QueueID)))

				// Then set routing rules if any
				if len(desiredBinding.RoutingRules) > 0 {
					log.Printf("Setting routing rules for binding: %s -> %s\n", exchangeID, desiredBinding.QueueID)
					err = client.UpdateMqBindingRoutingRules(orgID, envID, region, exchangeID, desiredBinding.QueueID, desiredBinding.RoutingRules)
					if err != nil {
						return fmt.Errorf("failed to set routing rules for binding %s -> %s: %v", exchangeID, desiredBinding.QueueID, err)
					}
					log.Println(color.Colorize(color.Green, fmt.Sprintf("Routing rules for [%s -> %s] successfully set", exchangeID, desiredBinding.QueueID)))
				}
			}
		} else if routingRulesNeedUpdate(desiredBinding.RoutingRules, existingBinding.RoutingRules) || viper.GetBool("force-update") {
			if dryRun {
				log.Println(color.Colorize(color.Yellow, fmt.Sprintf("[DRY-RUN] Would UPDATE routing rules for: [%s -> %s]", exchangeID, desiredBinding.QueueID)))
			} else {
				// Update routing rules
				log.Printf("Updating routing rules for binding: %s -> %s\n", exchangeID, desiredBinding.QueueID)
				err = client.UpdateMqBindingRoutingRules(orgID, envID, region, exchangeID, desiredBinding.QueueID, desiredBinding.RoutingRules)
				if err != nil {
					return fmt.Errorf("failed to update routing rules for binding %s -> %s: %v", exchangeID, desiredBinding.QueueID, err)
				}
				log.Println(color.Colorize(color.Green, fmt.Sprintf("Routing rules for [%s -> %s] successfully updated", exchangeID, desiredBinding.QueueID)))
			}
		} else {
			log.Println(color.Colorize(color.Blue, fmt.Sprintf("Binding [%s -> %s] already configured correctly", exchangeID, desiredBinding.QueueID)))
		}
	}

	// Note: We don't delete bindings that are no longer in the config - that's too dangerous.
	// Bindings must be deleted manually if needed.

	return nil
}

func routingRulesNeedUpdate(desired, current []anypointclient.MqRoutingRule) bool {
	if len(desired) != len(current) {
		return true
	}
	for i, desiredRule := range desired {
		currentRule := current[i]
		if desiredRule.PropertyName != currentRule.PropertyName {
			return true
		}
		if desiredRule.PropertyType != currentRule.PropertyType {
			return true
		}
		if desiredRule.MatcherType != currentRule.MatcherType {
			return true
		}
		if !reflect.DeepEqual(desiredRule.Value, currentRule.Value) {
			return true
		}
	}
	return false
}
