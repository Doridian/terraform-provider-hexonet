package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hexonet/go-sdk/v3/apiclient"
)

func resourceNameserver() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNameserverCreate,
		ReadContext:   resourceNameserverRead,
		UpdateContext: resourceNameserverUpdate,
		DeleteContext: resourceNameserverDelete,
		Schema:        makeNameserverSchema(false),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNameserverCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeNameserverCommand(cl, "AddNameserver", true, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return diags
}

func resourceNameserverRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return kindNameserverRead(ctx, d, m, false)
}

func resourceNameserverUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeNameserverCommand(cl, "ModifyNameserver", true, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return resourceDomainRead(ctx, d, m)
}

func resourceNameserverDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeNameserverCommand(cl, "DeleteNameserver", false, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return diags
}
