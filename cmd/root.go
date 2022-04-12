package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/Redpill-Linpro/anypointchdeployer/pkg/anypointclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "anypointchdeployer",
	Short: "CLI tool to deploy and update mule applications in Anypoint CloudHub",
	Long: `This is a command line tool to deploy new and update existing mule application running in Anypoint CloudHub using application artifacts stored in Exchange.
	
	Currently it can only deploy applications that are published to Exchange. There is no support for uploading and deploying local Mule application artifacts.`,
	Example:   "./chdeploy -u <username> -p <password> -o <organizationname> -e <environment> -a user  *.json",
	ValidArgs: []string{"*.json"},
	Run: func(cmd *cobra.Command, args []string) {
		deployConfig(args)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.ValidArgs = []string{"*.json"}
	rootCmd.Flags().StringP("region", "r", "US", "region for Anypoint. Use US for US control plane and EU for EU control plane. Or provide the base URL such as https://eu1.anypoint.mulesoft.com")
	rootCmd.Flags().StringP("authtype", "a", "connectedapp", "authentication method towards Anypoint Platform")
	rootCmd.Flags().StringP("bearer", "b", "", "authentication bearer token used to authenticate with Anypoint")
	rootCmd.Flags().StringP("user", "u", "", "user to use to login to Anypoint if token is not provided")
	rootCmd.Flags().StringP("password", "p", "", "password for the Anypoint user")
	rootCmd.Flags().StringP("client-id", "i", "", "client id for the Anypoint connected app")
	rootCmd.Flags().StringP("client-secret", "s", "", "client secret for the Anypoint connected app")
	rootCmd.Flags().StringP("organization", "o", "", "organization within Anypoint Platform")
	rootCmd.Flags().StringP("environment", "e", "", "environment within Anypoint Platform")
	rootCmd.Flags().IntP("concurrent-deployments", "c", 1, "max number of concurrent application deploys")
	viper.BindPFlag("region", rootCmd.Flags().Lookup("region"))
	viper.BindPFlag("authtype", rootCmd.Flags().Lookup("authtype"))
	viper.BindPFlag("bearer", rootCmd.Flags().Lookup("bearer"))
	viper.BindPFlag("user", rootCmd.Flags().Lookup("user"))
	viper.BindPFlag("password", rootCmd.Flags().Lookup("password"))
	viper.BindPFlag("client-id", rootCmd.Flags().Lookup("client-id"))
	viper.BindPFlag("client-secret", rootCmd.Flags().Lookup("client-secret"))
	viper.BindPFlag("organization", rootCmd.Flags().Lookup("organization"))
	viper.BindPFlag("environment", rootCmd.Flags().Lookup("environment"))
	viper.BindPFlag("concurrent-deployments", rootCmd.Flags().Lookup("concurrent-deployments"))
}

func deployConfig(args []string) {

	var wg sync.WaitGroup
	var newApplication anypointclient.CloudhubApplicationRequest
	guard := make(chan struct{}, viper.GetInt("concurrent-deployments"))
	client := getAnyPointClient()
	defer close(guard)

	for _, file := range args {
		wg.Add(1)
		go func(file string) {
			guard <- struct{}{}
			fmt.Printf("Reading file %s\n", file)

			f, _ := os.Open(file)
			defer f.Close()

			err := json.NewDecoder(f).Decode(&newApplication)
			if err != nil {
				log.Fatalf("Failed to decode %v %+v", flag.Arg(0), err)
			}

			organization, err := client.ResolveOrganization(viper.GetString("organization"))
			if err != nil {
				log.Fatalf("Failed to get organization %+v", err)
			}

			environment, err := client.ResolveEnvironment(organization, viper.GetString("environment"))
			if err != nil {
				log.Fatalf("Failed to get environment %+v", err)
			}

			application, err := client.GetApplication(environment, newApplication.ApplicationInfo.Domain)

			if err != nil {
				log.Fatalf("Failed to get application %+v", err)
			}

			if application.Domain == "" {
				err = client.CreateApplication(environment, newApplication)
				if err != nil {
					log.Fatalf("Failed to create applications %+v", err)
				}
				// TODO Check status
			} else if applicationHasChanged(application, newApplication) {
				err = client.UpdateApplication(environment, newApplication)
				if err != nil {
					log.Fatalf("Failed to create applications %+v", err)
				}
				// TODO Check status
			}
			fmt.Printf("Deployed %s\n", newApplication.ApplicationInfo.Domain)
			<-guard
		}(file)
	}
	wg.Wait()
	fmt.Printf("All done!\n")
}

func applicationHasChanged(application anypointclient.CloudhubApplicationResponse, newApplication anypointclient.CloudhubApplicationRequest) bool {
	// TODO: Add proper detection of changes to the deployment description
	return true
}

func getAnyPointClient() *anypointclient.AnypointClient {
	switch viper.GetString("authType") {
	case "bearer":
		if !viper.IsSet("bearer") {
			log.Println("Token must be supplied")
			os.Exit(10)
		}
		return anypointclient.NewAnypointClientWithToken(viper.GetString("region"), viper.GetString("bearer"))
	case "user":
		if !viper.IsSet("user") || !viper.IsSet("password") {
			log.Println("User and password must be supplied")
			os.Exit(10)
		}
		return anypointclient.NewAnypointClientWithCredentials(viper.GetString("region"), viper.GetString("user"), viper.GetString("password"))
	case "connectedapp":
		if !viper.IsSet("client-id") || !viper.IsSet("client-secret") {
			log.Println("Client id and secret must be supplied")
			os.Exit(10)
		}
		return anypointclient.NewAnypointClientWithConnectedApp(viper.GetString("region"), viper.GetString("client-id"), viper.GetString("client-secret"))
	default:
		log.Println("Authentication method must be supplied")
		os.Exit(10)
	}
	return nil
}
