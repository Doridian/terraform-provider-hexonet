package hexonet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

func makeContactSchema(readOnly bool) map[string]*schema.Schema {
	res := map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"first_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"middle_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"last_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"organization": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"address_line_1": {
			Type:     schema.TypeString,
			Required: true,
		},
		"address_line_2": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"city": {
			Type:     schema.TypeString,
			Required: true,
		},
		"state": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"zip": {
			Type:     schema.TypeString,
			Required: true,
		},
		"country": {
			Type:     schema.TypeString,
			Required: true,
		},
		"phone": {
			Type:     schema.TypeString,
			Required: true,
		},
		"fax": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"email": {
			Type:     schema.TypeString,
			Required: true,
		},
		"disclose": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"extra_attributes": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	if readOnly {
		makeSchemaReadOnly(res, "id")
	}

	return res
}

func makeContactCommand(cl *apiclient.APIClient, cmd CommandType, d *schema.ResourceData) *response.Response {
	req := map[string]interface{}{
		"COMMAND": fmt.Sprintf("%sContact", cmd),
	}

	if cmd == CommandCreate {
		req["NEW"] = "1"
	} else {
		id := d.Get("id").(string)
		if id == "" {
			id = d.Id()
		} else {
			d.SetId(id)
		}
		req["CONTACT"] = id
	}

	if cmd == CommandCreate || cmd == CommandUpdate {
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

		if cmd == CommandUpdate {
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

func kindContactRead(ctx context.Context, d *schema.ResourceData, m interface{}, addAll bool) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeContactCommand(cl, CommandRead, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	id := columnFirstOrDefault(resp, "ID", "").(string)
	d.SetId(id)
	if id == "" {
		return diags
	}

	d.Set("title", columnFirstOrDefault(resp, "TITLE", nil))
	d.Set("first_name", columnFirstOrDefault(resp, "FIRSTNAME", nil))
	d.Set("middle_name", columnFirstOrDefault(resp, "MIDDLENAME", nil))
	d.Set("last_name", columnFirstOrDefault(resp, "LASTNAME", nil))

	d.Set("organization", columnFirstOrDefault(resp, "ORGANIZATION", nil))

	d.Set("address_line_1", columnIndexOrDefault(resp, "STREET", nil, 0))
	d.Set("address_line_2", columnIndexOrDefault(resp, "STREET", nil, 1))

	d.Set("city", columnFirstOrDefault(resp, "CITY", nil))
	d.Set("state", columnFirstOrDefault(resp, "STATE", nil))
	d.Set("zip", columnFirstOrDefault(resp, "ZIP", nil))
	d.Set("country", columnFirstOrDefault(resp, "COUNTRY", nil))

	d.Set("phone", columnFirstOrDefault(resp, "PHONE", nil))
	d.Set("fax", columnFirstOrDefault(resp, "FAX", nil))
	d.Set("email", columnFirstOrDefault(resp, "EMAIL", nil))

	d.Set("disclose", numberStrToBool(columnFirstOrDefault(resp, "DISCLOSE", "0").(string)))

	handleExtraAttributesRead(d, resp, addAll)

	return diags
}
