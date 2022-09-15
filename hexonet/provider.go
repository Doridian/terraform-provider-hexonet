package hexonet

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

func New() provider.Provider {
	return &localProvider{}
}

type localProviderData struct {
	Username                types.String `tfsdk:"username"`
	Role                    types.String `tfsdk:"role"`
	Password                types.String `tfsdk:"password"`
	MfaToken                types.String `tfsdk:"mfa_token"`
	Live                    types.Bool   `tfsdk:"live"`
	HighPerformance         types.Bool   `tfsdk:"high_performance"`
	AllowDomainCreateDelete types.Bool   `tfsdk:"allow_domain_create_delete"`
}

type localProvider struct {
	allowDomainCreateDelete bool
	configured              bool
	client                  *apiclient.APIClient
}

func envVarForKey(key string) string {
	return fmt.Sprintf("HEXONET_%s", strings.ToUpper(key))
}

func envDescription(desc, key string) string {
	return fmt.Sprintf("%s (environment variable %s)", desc, envVarForKey(key))
}

func (p *localProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"username": {
				Type:        types.StringType,
				Optional:    true,
				Description: envDescription("Username", "username"),
			},
			"role": {
				Type:        types.StringType,
				Optional:    true,
				Description: envDescription("Role (sub-user)", "role"),
			},
			"password": {
				Type:        types.StringType,
				Sensitive:   true,
				Optional:    true,
				Description: envDescription("Password", "password"),
			},
			"mfa_token": {
				Type:        types.StringType,
				Sensitive:   true,
				Optional:    true,
				Description: envDescription("MFA token (required if MFA is enabled)", "mfa_token"),
			},
			"live": {
				Type:        types.BoolType,
				Optional:    true,
				Description: envDescription("Whether to use the live (true) or the OTE/test (false) system", "live"),
			},
			"high_performance": {
				Type:        types.BoolType,
				Optional:    true,
				Description: envDescription("Whether to use high-performance connection establishment (might need additional setup)", "high_performance"),
			},
			"allow_domain_create_delete": {
				Type:        types.BoolType,
				Required:    true,
				Description: "Whether to use AddDomain / DeleteDomain to send domain registration/deletion requests, otherwise will only read and update domains, never register or delete (extreme caution should be taken when enabling this option!)",
			},
		},
		Description: "Provider for Hexonet domain API",
	}, nil
}

func (p *localProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hexonet"
}

func (p *localProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newResourceContact,
		newResourceDomain,
		newResourceNameServer,
	}
}

func (p *localProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newDataSourceContact,
		newDataSourceDomain,
		newDataSourceNameServer,
	}
}

func getValueOrDefaultToEnv(val types.String, key string, resp *provider.ConfigureResponse, allowEmpty bool) string {
	if val.Unknown {
		resp.Diagnostics.AddError("Can not configure client", fmt.Sprintf("Unknown value for %s", key))
		return ""
	}

	var res string
	if val.Null {
		res = os.Getenv(envVarForKey(key))
	} else {
		res = val.Value
	}

	if res == "" && !allowEmpty {
		resp.Diagnostics.AddError("Can not configure client", fmt.Sprintf("Empty value for %s", key))
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

	if !config.AllowDomainCreateDelete.Null && !config.AllowDomainCreateDelete.Unknown {
		p.allowDomainCreateDelete = config.AllowDomainCreateDelete.Value
	} else {
		p.allowDomainCreateDelete = false
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

	utils.HandlePossibleErrorResponse(res, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	p.client = c
	p.configured = true

	resp.DataSourceData = p
	resp.ResourceData = p
}
