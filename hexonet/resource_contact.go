package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hexonet/go-sdk/v3/apiclient"
)

func resourceContact() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContactCreate,
		ReadContext:   resourceContactRead,
		UpdateContext: resourceContactUpdate,
		DeleteContext: resourceContactDelete,
		Schema:        makeContactSchema(false),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceContactCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeContactCommand(cl, "AddContact", true, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	id := columnFirstOrDefault(resp, "CONTACT", nil).(string)
	d.SetId(id)

	return diags
}

func resourceContactRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return kindContactRead(ctx, d, m, false)
}

func resourceContactUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeContactCommand(cl, "ModifyContact", true, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return resourceContactRead(ctx, d, m)
}

func resourceContactDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeContactCommand(cl, "DeleteContact", false, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return diags
}
