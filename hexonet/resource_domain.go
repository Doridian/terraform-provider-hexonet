package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainCreate,
		ReadContext:   resourceDomainRead,
		UpdateContext: resourceDomainUpdate,
		DeleteContext: resourceDomainDelete,
		Schema:        makeDomainSchema(false),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func makeDomainCommand(cl *apiclient.APIClient, cmd string, addData bool, d *schema.ResourceData) *response.Response {
	domain := d.Get("domain").(string)
	if domain == "" {
		domain = d.Id()
	} else {
		d.SetId(domain)
	}

	req := map[string]interface{}{
		"COMMAND": cmd,
		"DOMAIN":  domain,
	}

	if addData {
		fillRequestArray(d.Get("name_servers").([]interface{}), "NAMESERVER", req, MAX_NAMESERVERS, false)

		fillRequestArray(d.Get("owner_contacts").([]interface{}), "OWNERCONTACT", req, 1, false)
		fillRequestArray(d.Get("admin_contacts").([]interface{}), "ADMINCONTACT", req, MAX_CONTACTS, false)
		fillRequestArray(d.Get("tech_contacts").([]interface{}), "TECHCONTACT", req, MAX_CONTACTS, false)
		fillRequestArray(d.Get("billing_contacts").([]interface{}), "BILLINGCONTACT", req, MAX_CONTACTS, true)

		req["TRANSFERLOCK"] = boolToNumberStr(d.Get("transfer_lock").(bool))

		handleExtraAttributesWrite(d, req)
	}

	return cl.Request(req)
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeDomainCommand(cl, "AddDomain", true, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return diags
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return kindDomainRead(ctx, d, m, false)
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeDomainCommand(cl, "ModifyDomain", true, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return resourceDomainRead(ctx, d, m)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeDomainCommand(cl, "DeleteDomain", false, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return diags
}
