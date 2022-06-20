package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/Redpill-Linpro/anypointchdeployer/internal/appconf"
	"github.com/Redpill-Linpro/anypointchdeployer/internal/flagvalidator"
	"github.com/Redpill-Linpro/anypointchdeployer/pkg/anypointclient"
	"github.com/TwiN/go-color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		if err := flagvalidator.ValidateFlags(); err != nil {
			log.Fatal(err)
		}
		client := appconf.GetAnypointClient()
		deployConfig(client, args)
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
	rootCmd.Flags().StringP("authtype", "a", "connectedapp", "authentication method towards Anypoint Platform")
	rootCmd.Flags().StringP("bearer", "b", "", "authentication bearer token used to authenticate with Anypoint")
	rootCmd.Flags().StringP("user", "u", "", "user to use to login to Anypoint if token is not provided")
	rootCmd.Flags().StringP("password", "p", "", "password for the Anypoint user")
	rootCmd.Flags().StringP("client-id", "i", "", "client id for the Anypoint connected app")
	rootCmd.Flags().StringP("client-secret", "s", "", "client secret for the Anypoint connected app")
	rootCmd.Flags().StringP("organization", "o", "", "organization within Anypoint Platform")
	rootCmd.Flags().StringP("environment", "e", "", "environment within Anypoint Platform")
	rootCmd.Flags().IntP("concurrent-deployments", "c", 1, "max number of concurrent application deploys")
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(f.Name, f)
	})
	flagvalidator.AddFlagSetValidator("region", []any{"US", "EU"})
	flagvalidator.AddFlagSetValidator("authType", []any{"bearer", "user", "connectedapp"})
	flagvalidator.AddFlagSetValidator("concurrent-deployments", []any{1, 2, 3, 4, 5})
}

func deployConfig(client *anypointclient.AnypointClient, files []string) {

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

			var newApplication anypointclient.CloudhubApplicationRequest

			log.Printf("Reading file: %s", file)

			f, _ := os.Open(file)
			defer f.Close()

			if err := json.NewDecoder(f).Decode(&newApplication); err != nil {
				faults <- fmt.Errorf("Failed to decode %v %+v", flag.Arg(0), err)
				return
			}

			appconf.AddApplicationGav(newApplication)

			organization, err := client.ResolveOrganization(viper.GetString("organization"))
			if err != nil {
				faults <- fmt.Errorf("Failed to get organization %+v", err)
				return
			}

			environment, err := client.ResolveEnvironment(organization, viper.GetString("environment"))
			if err != nil {
				faults <- fmt.Errorf("Failed to get environment %+v", err)
				return
			}
			application, err := client.GetApplication(environment, newApplication.ApplicationInfo.Domain)
			if err != nil {
				faults <- fmt.Errorf("Failed to get application %+v", err)
				return
			}
			if application.Domain == "" {
				if err = client.CreateApplication(environment, newApplication); err != nil {
					faults <- fmt.Errorf("Failed to create application %+v", err)
					return
				}
				// TODO Check status
			} else if appconf.ApplicationHasChanged(application, newApplication) {
				if err := client.UpdateApplication(environment, newApplication); err != nil {
					faults <- fmt.Errorf("Failed to update application: %s\nCause: %+v", newApplication.ApplicationSource.ArtifactID, err)
					return
				}
				// TODO Check status
			}
			log.Println(color.Colorize(color.Green, fmt.Sprintf("Application: [%s] successfully deployed", application.Domain)))
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
	log.Println(color.Colorize(color.Green, fmt.Sprintf("All updated applications deployed successfully!\n")))
}
