package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type resourceNameServerType struct{}

type resourceNameServer struct {
	p localProvider
}

func (r resourceNameServerType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: makeNameServerSchema(false),
	}, nil
}

func (r resourceNameServerType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceNameServer{
		p: *(p.(*localProvider)),
	}, nil
}

func (r resourceNameServer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var data NameServer
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeNameServerCommand(r.p.client, CommandCreate, data, NameServer{}, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindNameserverRead(ctx, data, r.p.client, false, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceNameServer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var data NameServer
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindNameserverRead(ctx, data, r.p.client, false, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceNameServer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var data NameServer
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var dataOld NameServer
	diags = req.State.Get(ctx, &dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeNameServerCommand(r.p.client, CommandUpdate, data, dataOld, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindNameserverRead(ctx, data, r.p.client, false, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceNameServer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var dataOld NameServer
	diags := req.State.Get(ctx, &dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeNameServerCommand(r.p.client, CommandDelete, NameServer{
		NameServer: dataOld.NameServer,
	}, dataOld, resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.RemoveResource(ctx)
}
