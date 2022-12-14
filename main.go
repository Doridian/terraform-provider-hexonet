package main

import (
	"context"

	"github.com/Doridian/terraform-provider-hexonet/hexonet"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	providerserver.Serve(context.Background(), hexonet.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/doridian/hexonet",
	})
}
