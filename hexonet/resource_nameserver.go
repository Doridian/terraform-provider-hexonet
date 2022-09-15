package hexonet

import (
	"context"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type resourceNameServer struct {
	p *localProvider
}

func newResourceNameServer() resource.Resource {
	return &resourceNameServer{}
}

func (r *resourceNameServer) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:  makeNameServerSchema(false),
		Description: "Nameserver object, used to register so-called \"glue\" records when a domain's nameservers use hosts on the same domain (example: example.com using ns1.example.com)",
	}, nil
}

func (r *resourceNameServer) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.p = req.ProviderData.(*localProvider)
}

func (r *resourceNameServer) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameserver"
}

func (r resourceNameServer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &NameServer{}
	diags := req.Plan.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeNameServerCommand(ctx, r.p.client, utils.CommandCreate, data, &NameServer{}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindNameserverRead(ctx, data, r.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceNameServer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &NameServer{}
	diags := req.State.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindNameserverRead(ctx, data, r.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceNameServer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &NameServer{}
	diags := req.Plan.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataOld := &NameServer{}
	diags = req.State.Get(ctx, dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeNameServerCommand(ctx, r.p.client, utils.CommandUpdate, data, dataOld, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindNameserverRead(ctx, data, r.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, data)
}

func (r resourceNameServer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	dataOld := &NameServer{}
	diags := req.State.Get(ctx, &dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeNameServerCommand(ctx, r.p.client, utils.CommandDelete, &NameServer{
		Host: dataOld.Host,
	}, dataOld, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r resourceNameServer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("host"), req, resp)
}
