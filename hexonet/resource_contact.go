package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type resourceContactType struct{}

type resourceContact struct {
	p localProvider
}

func (r resourceContactType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: makeContactSchema(false),
	}, nil
}

func (r resourceContactType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceContact{
		p: *(p.(*localProvider)),
	}, nil
}

func (r resourceContact) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var data Contact
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeContactCommand(r.p.client, CommandCreate, data, Contact{}, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindContactRead(ctx, data, r.p.client, false, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceContact) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var data Contact
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindContactRead(ctx, data, r.p.client, false, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceContact) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var data Contact
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var dataOld Contact
	diags = req.State.Get(ctx, &dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeContactCommand(r.p.client, CommandUpdate, data, dataOld, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindContactRead(ctx, data, r.p.client, false, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceContact) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
		makeNotConfiguredError(&resp.Diagnostics)
		return
	}

	var dataOld Contact
	diags := req.State.Get(ctx, &dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeContactCommand(r.p.client, CommandDelete, Contact{
		ID: dataOld.ID,
	}, dataOld, resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r resourceContact) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
