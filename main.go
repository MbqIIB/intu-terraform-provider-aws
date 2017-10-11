package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/intu-terraform-provider-aws/aws"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: aws.Provider})
}
