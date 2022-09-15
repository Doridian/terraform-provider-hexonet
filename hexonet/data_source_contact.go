package hexonet

import (
	"context"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type dataSourceContact struct {
	p *localProvider
}

func newDataSourceContact() datasource.DataSource {
	return &dataSourceContact{}
}

func (d *dataSourceContact) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:  makeContactSchema(true),
		Description: "Contact object, used for domain owner/admin/...",
	}, nil
}

func (d *dataSourceContact) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	resp.State.Set(ctx, data)
}
