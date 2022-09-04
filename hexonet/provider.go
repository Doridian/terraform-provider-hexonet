package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"high_performance": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEXONET_HIGH_PERFORMANCE", false),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEXONET_USERNAME", nil),
			},
			"role": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEXONET_ROLE", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEXONET_PASSWORD", nil),
			},
			"mfa_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("HEXONET_MFA_TOKEN", nil),
			},
			"live": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEXONET_LIVE", false),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hexonet_domain":     resourceDomain(),
			"hexonet_nameserver": resourceNameserver(),
			"hexonet_contact":    resourceContact(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hexonet_domain":     dataSourceDomain(),
			"hexonet_nameserver": dataSourceNameserver(),
			"hexonet_contact":    dataSourceContact(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	role := d.Get("role").(string)
	mfatoken := d.Get("mfa_token").(string)
	highperformance := d.Get("high_performance").(bool)
	live := d.Get("live").(bool)

	var diags diag.Diagnostics

	c := apiclient.NewAPIClient()
	if live {
		c.UseLIVESystem()
	} else {
		c.UseOTESystem()
	}

	if highperformance {
		c.UseHighPerformanceConnectionSetup()
	} else {
		c.UseDefaultConnectionSetup()
	}

	if role != "" {
		c.SetRoleCredentials(username, role, password)
	} else {
		c.SetCredentials(username, password)
	}

	var res *response.Response
	if mfatoken != "" {
		res = c.Login()
	} else {
		res = c.Login(mfatoken)
	}

	if res.IsError() {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to log into Hexonet API",
			Detail:   res.Raw,
		})
		return nil, diags
	}

	return c, diags
}
