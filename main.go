package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/curve-technology/terraform-provider-confluent-schema-registry/pkg/tf"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: tf.Provider,
	})
}
