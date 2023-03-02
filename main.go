package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/OJFord/terraform-provider-mullvad/provider"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(
		context.Background(),
		provider.New,
		providerserver.ServeOpts{
			Address: "registry.terraform.io/OJFord/mullvad",
			Debug:   debug,
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
