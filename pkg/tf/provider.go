package tf

import (
	"context"

	"github.com/riferrei/srclient"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "",
				Sensitive: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"schemaregistry_bulk_dependent_schemas": resourceBulkDependentSchemas(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	url := d.Get("url").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client := srclient.CreateSchemaRegistryClient(url)
	client.CachingEnabled(false)

	if (username != "") && (password != "") {
		client.SetCredentials(username, password)
	}

	return client, diags
}
