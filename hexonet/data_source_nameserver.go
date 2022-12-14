package hexonet

import (
	"context"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type dataSourceNameServer struct {
	p *localProvider
}

func newDataSourceNameServer() datasource.DataSource {
	return &dataSourceNameServer{}
}

func (r *dataSourceNameServer) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:  utils.ResourceSchemaToDataSourceSchema(makeNameServerResourceSchema(), "host"),
		Description: "Nameserver object, used to register so-called \"glue\" records when a domain's nameservers use hosts on the same domain (example: example.com using ns1.example.com)",
	}
}

func (d *dataSourceNameServer) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.p = req.ProviderData.(*localProvider)
}

func (d *dataSourceNameServer) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameserver"
}

func (d *dataSourceNameServer) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &NameServer{}
	diags := req.Config.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindNameserverRead(ctx, data, d.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}
