package hexonet

import (
	"context"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type dataSourceContact struct {
	p *localProvider
}

func newDataSourceContact() datasource.DataSource {
	return &dataSourceContact{}
}

func (r *dataSourceContact) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:  utils.ResourceSchemaToDataSourceSchema(makeContactResourceSchema(), "id"),
		Description: "Contact object, used for domain owner/admin/...",
	}
}

func (d *dataSourceContact) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.p = req.ProviderData.(*localProvider)
}

func (d *dataSourceContact) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact"
}

func (d *dataSourceContact) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &Contact{}
	diags := req.Config.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindContactRead(ctx, data, d.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
