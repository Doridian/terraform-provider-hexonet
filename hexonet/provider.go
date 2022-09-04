package hexonet

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

func New() provider.Provider {
	return &localProvider{}
}

type localProvider struct {
	configured bool
	client     *apiclient.APIClient
}

func (p *localProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"username": {
				Type:     types.StringType,
				Optional: true,
			},
			"role": {
				Type:     types.StringType,
				Optional: true,
			},
			"password": {
				Type:      types.StringType,
				Sensitive: true,
				Optional:  true,
			},
			"mfa_token": {
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
			},
			"live": {
				Type:     types.BoolType,
				Optional: true,
			},
			"high_performance": {
				Type:     types.BoolType,
				Optional: true,
			},
		},
	}, nil
}

type localProviderData struct {
	Username        types.String `tfsdk:"username"`
	Role            types.String `tfsdk:"role"`
	Password        types.String `tfsdk:"password"`
	MfaToken        types.String `tfsdk:"mfa_token"`
	Live            types.Bool   `tfsdk:"live"`
	HighPerformance types.Bool   `tfsdk:"high_performance"`
}

func (p *localProvider) GetResources(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{
		"hexonet_domain":     resourceDomainType{},
		"hexonet_nameserver": resourceNameServerType{},
		"hexonet_contact":    resourceContactType{},
	}, nil
}

func (p *localProvider) GetDataSources(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{
		"hexonet_domain":     dataSourceDomainType{},
		"hexonet_nameserver": dataSourceNameServerType{},
		"hexonet_contact":    dataSourceContactType{},
	}, nil
}

func getValueOrDefaultToEnv(val types.String, env string, resp *provider.ConfigureResponse, allowEmpty bool) string {
	if val.Unknown {
		resp.Diagnostics.AddError("Can not configure client", fmt.Sprintf("Unknown value for %s", env))
		return ""
	}

	var res string
	if val.Null {
		res = os.Getenv(fmt.Sprintf("HEXONET_%s", strings.ToUpper(env)))
	} else {
		res = val.Value
	}

	if res == "" && !allowEmpty {
		resp.Diagnostics.AddError("Can not configure client", fmt.Sprintf("Empty value for %s", env))
	}
	return res
}

func (p *localProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config localProviderData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := getValueOrDefaultToEnv(config.Username, "username", resp, false)
	password := getValueOrDefaultToEnv(config.Password, "password", resp, false)
	role := getValueOrDefaultToEnv(config.Role, "role", resp, true)
	mfaToken := getValueOrDefaultToEnv(config.MfaToken, "mfa_token", resp, true)

	highPerformance := false
	live := true

	if !config.HighPerformance.Null && !config.HighPerformance.Unknown {
		highPerformance = config.HighPerformance.Value
	}

	if !config.Live.Null && !config.Live.Unknown {
		live = config.Live.Value
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c := apiclient.NewAPIClient()
	if live {
		c.UseLIVESystem()
	} else {
		c.UseOTESystem()
	}

	if highPerformance {
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
	if mfaToken != "" {
		res = c.Login()
	} else {
		res = c.Login(mfaToken)
	}

	handlePossibleErrorResponse(res, resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	p.client = c
	p.configured = true
}
