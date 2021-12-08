package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Redpill-Linpro/anypointchdeployer/pkg/anypointclient"
)

func main() {
	regionPtr := flag.String("region", "US", "region for Anypoint. Use US for US control plane and EU for EU control plane. Or provide the base URL such as https://eu1.anypoint.mulesoft.com")
	bearerPtr := flag.String("bearer", "--EMPTY--", "authentication bearer token used to authenticate with Anypoint")
	userPtr := flag.String("user", "--EMPTY--", "user to use to login to Anypoint if token is not provided")
	passwordPtr := flag.String("password", "--EMPTY--", "password for the Anypoint user")
	organizationPtr := flag.String("organization", "--EMPTY--", "organization within Anypoint Platform")
	environmentPtr := flag.String("environment", "--EMPTY--", "environment within Anypoint Platform")

	flag.Parse()

	var client *anypointclient.AnypointClient

	if *bearerPtr == "--EMPTY--" {
		if *userPtr == "--EMPTY--" || *passwordPtr == "--EMPTY--" {
			log.Println("Token or user and password must be supplied")
			os.Exit(10)
		}
		client = anypointclient.NewAnypointClientWithCredentials(*regionPtr,
			*userPtr, *passwordPtr)
	} else {
		client = anypointclient.NewAnypointClientWithToken(*regionPtr, *bearerPtr)
	}

	var newApplication anypointclient.CloudhubApplicationRequest
	for _, file := range flag.Args() {

		fmt.Printf("Reading file %s\n", file)

		f, _ := os.Open(file)
		defer f.Close()

		err := json.NewDecoder(f).Decode(&newApplication)
		if err != nil {
			log.Fatalf("Failed to decode %v %+v", flag.Arg(0), err)
		}
		organisation, err := client.ResolveOrganisation(*organizationPtr)
		if err != nil {
			log.Fatalf("Failed to get organisation %+v", err)
		}
		environment, err := client.ResolveEnvironment(organisation, *environmentPtr)
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

	}
	fmt.Printf("All done!\n")

}

func applicationHasChanged(application anypointclient.CloudhubApplicationResponse, newApplication anypointclient.CloudhubApplicationRequest) bool {
        // TODO: Add proper detection of changes to the deployment description
	return true
}
