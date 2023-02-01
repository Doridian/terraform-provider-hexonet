package hexonet

import (
	"context"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type resourceDomain struct {
	p *localProvider
}

func newResourceDomain() resource.Resource {
	return &resourceDomain{}
}

func (r *resourceDomain) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:  makeDomainResourceSchema(),
		Description: "Domain object, can be used to configure most attributes of domains",
	}
}

func (r *resourceDomain) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*localProvider)
}

func (r *resourceDomain) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *resourceDomain) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &Domain{}
	diags := req.Plan.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.p.allowDomainCreateDelete {
		_ = makeDomainCommand(ctx, r.p.client, utils.CommandCreate, data, &Domain{}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	data = kindDomainRead(ctx, data, r.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceDomain) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &Domain{}
	diags := req.State.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindDomainRead(ctx, data, r.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceDomain) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	data := &Domain{}
	diags := req.Plan.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataOld := &Domain{}
	diags = req.State.Get(ctx, dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = makeDomainCommand(ctx, r.p.client, utils.CommandUpdate, data, dataOld, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data = kindDomainRead(ctx, data, r.p.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceDomain) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
		utils.MakeNotConfiguredError(&resp.Diagnostics)
		return
	}

	dataOld := &Domain{}
	diags := req.State.Get(ctx, dataOld)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.p.allowDomainCreateDelete {
		_ = makeDomainCommand(ctx, r.p.client, utils.CommandDelete, &Domain{
			Domain: dataOld.Domain,
		}, dataOld, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *resourceDomain) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}
