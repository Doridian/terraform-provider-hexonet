package hexonet

import (
	"context"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type resourceContact struct {
	p *localProvider
}

func newResourceContact() resource.Resource {
	return &resourceContact{}
}

func (r *resourceContact) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:  makeContactSchema(false),
		Description: "Contact object, used for domain owner/admin/...",
	}, nil
}

func (r *resourceContact) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*localProvider)
}

func (r *resourceContact) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact"
}

func (r *resourceContact) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &Contact{}
	diags := req.Plan.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeContactCommand(ctx, r.p.client, utils.CommandCreate, data, &Contact{}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindContactRead(ctx, data, r.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r *resourceContact) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &Contact{}
	diags := req.State.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindContactRead(ctx, data, r.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r *resourceContact) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &Contact{}
	diags := req.Plan.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataOld := &Contact{}
	diags = req.State.Get(ctx, dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeContactCommand(ctx, r.p.client, utils.CommandUpdate, data, dataOld, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindContactRead(ctx, data, r.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r *resourceContact) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	dataOld := &Contact{}
	diags := req.State.Get(ctx, dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeContactCommand(ctx, r.p.client, utils.CommandDelete, &Contact{
		ID: dataOld.ID,
	}, dataOld, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *resourceContact) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
