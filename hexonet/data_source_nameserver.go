package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type dataSourceNameServerType struct{}

type dataSourceNameServer struct {
	p localProvider
}

func (d dataSourceNameServerType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: makeNameServerSchema(true),
	}, nil
}

func (d dataSourceNameServerType) NewDataSource(_ context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return dataSourceNameServer{
		p: *(p.(*localProvider)),
	}, nil
}

func (d dataSourceNameServer) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var data NameServer
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindNameserverRead(ctx, data, d.p.client, false, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}
