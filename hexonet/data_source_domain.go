package hexonet

import (
	"context"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type dataSourceDomainType struct{}

type dataSourceDomain struct {
	p localProvider
}

func (d dataSourceDomainType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: makeDomainSchema(true),
	}, nil
}

func (d dataSourceDomainType) NewDataSource(_ context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return dataSourceDomain{
		p: *(p.(*localProvider)),
	}, nil
}

func (d dataSourceDomain) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var data Domain
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindDomainRead(ctx, data, d.p.client, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}
