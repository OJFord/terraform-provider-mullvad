package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/OJFord/terraform-provider-mullvad/provider"
)

func main() {
	err := providerserver.Serve(
		context.Background(),
		provider.New,
		providerserver.ServeOpts{
			Address: "registry.terraform.io/OJFord/mullvad",
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
