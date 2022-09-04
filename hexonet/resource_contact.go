package hexonet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
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

func makeContactCommand(cl *apiclient.APIClient, cmd string, addData bool, d *schema.ResourceData) *response.Response {
	req := map[string]interface{}{
		"COMMAND": cmd,
	}

	isAdd := cmd == "AddContact"
	isModify := cmd == "ModifyContact"

	if isAdd {
		req["NEW"] = "1"
	} else {
		req["CONTACT"] = d.Id()
	}

	if addData {
		optionals := []string{"TITLE", "MIDDLENAME", "ORGANIZATION", "STATE", "FAX"}

		req["TITLE"] = d.Get("title").(string)
		req["FIRSTNAME"] = d.Get("first_name").(string)
		req["MIDDLENAME"] = d.Get("middle_name").(string)
		req["LASTNAME"] = d.Get("last_name").(string)

		req["ORGANIZATION"] = d.Get("organization").(string)

		req["STREET0"] = d.Get("address_line_1").(string)
		req["STREET1"] = d.Get("address_line_2").(string)
		req["CITY"] = d.Get("city").(string)
		req["STATE"] = d.Get("state").(string)
		req["ZIP"] = d.Get("zip").(string)
		req["COUNTRY"] = d.Get("country").(string)

		req["PHONE"] = d.Get("phone").(string)
		req["FAX"] = d.Get("fax").(string)
		req["EMAIL"] = d.Get("email").(string)

		req["DISCLOSE"] = boolToNumberStr(d.Get("disclose").(bool))

		if isModify {
			i := 0
			for _, optional := range optionals {
				if req[optional] != "" {
					continue
				}

				req[fmt.Sprintf("DELETE%d", i)] = optional
				delete(req, optional)
				i++
			}
		}

		handleExtraAttributesWrite(d, req)
	}

	return cl.Request(req)
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
