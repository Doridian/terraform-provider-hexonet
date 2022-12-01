package hexonet

import (
	"context"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type dataSourceDomain struct {
	p *localProvider
}

func newDataSourceDomain() datasource.DataSource {
	return &dataSourceDomain{}
}

func (r *dataSourceDomain) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:  makeDomainSchema(true),
		Description: "Domain object",
	}
}

func (d *dataSourceDomain) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.p = req.ProviderData.(*localProvider)
}

func (d *dataSourceDomain) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *dataSourceDomain) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &Domain{}
	diags := req.Config.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindDomainRead(ctx, data, d.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}
