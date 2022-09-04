package hexonet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

func makeContactSchema(readOnly bool) map[string]tfsdk.Attribute {
	res := map[string]tfsdk.Attribute{
		"id": {
			Type:     types.StringType,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.RequiresReplace(),
				resource.UseStateForUnknown(),
			},
		},
		"title": {
			Type:     types.StringType,
			Optional: true,
		},
		"first_name": {
			Type:     types.StringType,
			Required: true,
		},
		"middle_name": {
			Type:     types.StringType,
			Optional: true,
		},
		"last_name": {
			Type:     types.StringType,
			Required: true,
		},
		"organization": {
			Type:     types.StringType,
			Optional: true,
		},
		"address_line_1": {
			Type:     types.StringType,
			Required: true,
		},
		"address_line_2": {
			Type:     types.StringType,
			Optional: true,
		},
		"city": {
			Type:     types.StringType,
			Required: true,
		},
		"state": {
			Type:     types.StringType,
			Optional: true,
		},
		"zip": {
			Type:     types.StringType,
			Required: true,
		},
		"country": {
			Type:     types.StringType,
			Required: true,
		},
		"phone": {
			Type:     types.StringType,
			Required: true,
		},
		"fax": {
			Type:     types.StringType,
			Optional: true,
		},
		"email": {
			Type:     types.StringType,
			Required: true,
		},
		"disclose": {
			Type:     types.BoolType,
			Required: true,
		},
		"extra_attributes": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
	}

	if readOnly {
		makeSchemaReadOnly(res, "id")
	}

	return res
}

type Contact struct {
	ID types.String `tfsdk:"id"`

	Title      types.String `tfsdk:"title"`
	FirstName  types.String `tfsdk:"first_name"`
	MiddleName types.String `tfsdk:"middle_name"`
	LastName   types.String `tfsdk:"last_name"`

	Organization types.String `tfsdk:"organization"`

	AddressLine1 types.String `tfsdk:"address_line_1"`
	AddressLine2 types.String `tfsdk:"address_line_2"`

	City   types.String `tfsdk:"city"`
	State  types.String `tfsdk:"state"`
	ZIP    types.String `tfsdk:"zip"`
	Coutry types.String `tfsdk:"country"`

	Phone types.String `tfsdk:"phone"`
	Fax   types.String `tfsdk:"fax"`
	Email types.String `tfsdk:"email"`

	Disclose types.Bool `tfsdk:"disclose"`

	ExtraAttributes types.Map `tfsdk:"extra_attributes"`
}

func makeContactCommand(cl *apiclient.APIClient, cmd CommandType, contact Contact, oldContact Contact, diag diag.Diagnostics) *response.Response {
	req := map[string]interface{}{
		"COMMAND": fmt.Sprintf("%sContact", cmd),
	}

	if cmd == CommandCreate {
		req["NEW"] = "1"
	} else {
		if contact.ID.Null || contact.ID.Unknown {
			diag.AddError("Main ID attribute unknwon or null", "id is null or unknown")
			return nil
		}

		if !oldContact.ID.Null && !oldContact.ID.Unknown && oldContact.ID.Value != contact.ID.Value {
			diag.AddError("Main ID attribute changed", fmt.Sprintf("id changed from %s to %s", oldContact.ID.Value, contact.ID.Value))
			return nil
		}

		req["CONTACT"] = contact.ID.Value
	}

	if cmd == CommandCreate || cmd == CommandUpdate {
		optionals := []string{"TITLE", "MIDDLENAME", "ORGANIZATION", "STATE", "FAX"}

		req["TITLE"] = autoUnboxString(contact.Title, "")
		req["FIRSTNAME"] = autoUnboxString(contact.FirstName, "")
		req["MIDDLENAME"] = autoUnboxString(contact.MiddleName, "")
		req["LASTNAME"] = autoUnboxString(contact.LastName, "")

		req["ORGANIZATION"] = autoUnboxString(contact.Organization, "")

		req["STREET0"] = autoUnboxString(contact.AddressLine1, "")
		req["STREET1"] = autoUnboxString(contact.AddressLine2, "")
		req["CITY"] = autoUnboxString(contact.City, "")
		req["STATE"] = autoUnboxString(contact.State, "")
		req["ZIP"] = autoUnboxString(contact.ZIP, "")
		req["COUNTRY"] = autoUnboxString(contact.Coutry, "")

		req["PHONE"] = autoUnboxString(contact.Phone, "")
		req["FAX"] = autoUnboxString(contact.Fax, "")
		req["EMAIL"] = autoUnboxString(contact.Email, "")

		if !contact.Disclose.Null && !contact.Disclose.Unknown {
			req["DISCLOSE"] = boolToNumberStr(contact.Disclose.Value)
		}

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

		handleExtraAttributesWrite(contact.ExtraAttributes, oldContact.ExtraAttributes, req)
	}

	return cl.Request(req)
}

func kindContactRead(ctx context.Context, contact Contact, cl *apiclient.APIClient, diag diag.Diagnostics) Contact {
	resp := makeContactCommand(cl, CommandRead, contact, contact, diag)
	if diag.HasError() {
		return Contact{}
	}

	return Contact{
		ID: types.String{Value: columnFirstOrDefault(resp, "ID", "").(string)},

		Title:      autoBoxString(columnFirstOrDefault(resp, "TITLE", nil)),
		FirstName:  autoBoxString(columnFirstOrDefault(resp, "FIRSTNAME", nil)),
		MiddleName: autoBoxString(columnFirstOrDefault(resp, "MIDDLENAME", nil)),
		LastName:   autoBoxString(columnFirstOrDefault(resp, "LASTNAME", nil)),

		Organization: autoBoxString(columnFirstOrDefault(resp, "ORGANIZATION", nil)),

		AddressLine1: autoBoxString(columnIndexOrDefault(resp, "STREET", nil, 0)),
		AddressLine2: autoBoxString(columnIndexOrDefault(resp, "STREET", nil, 1)),
		City:         autoBoxString(columnFirstOrDefault(resp, "CITY", nil)),
		ZIP:          autoBoxString(columnFirstOrDefault(resp, "ZIP", nil)),
		State:        autoBoxString(columnFirstOrDefault(resp, "STATE", nil)),
		Coutry:       autoBoxString(columnFirstOrDefault(resp, "COUNTRY", nil)),

		Phone: autoBoxString(columnFirstOrDefault(resp, "PHONE", nil)),
		Fax:   autoBoxString(columnFirstOrDefault(resp, "FAX", nil)),
		Email: autoBoxString(columnFirstOrDefault(resp, "EMAIL", nil)),

		Disclose: autoBoxBoolNumberStr(columnFirstOrDefault(resp, "DISCLOSE", "0")),

		ExtraAttributes: handleExtraAttributesRead(resp),
	}
}
