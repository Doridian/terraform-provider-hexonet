package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type resourceDomainType struct{}

type resourceDomain struct {
	p localProvider
}

func (r resourceDomainType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: makeDomainSchema(false),
	}, nil
}

func (r resourceDomainType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceDomain{
		p: *(p.(*localProvider)),
	}, nil
}

func (r resourceDomain) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

func (r resourceDomain) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var data Domain
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindDomainRead(ctx, data, r.p.client, false, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceDomain) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var data Domain
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var dataOld Domain
	diags = req.State.Get(ctx, &dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeDomainCommand(r.p.client, CommandUpdate, data, dataOld, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindDomainRead(ctx, data, r.p.client, false, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceDomain) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var dataOld Domain
	diags := req.State.Get(ctx, &dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeDomainCommand(r.p.client, CommandDelete, Domain{
		Domain: dataOld.Domain,
	}, dataOld, resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r resourceDomain) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}
