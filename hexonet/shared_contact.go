package hexonet

import (
	"context"
	"fmt"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
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

func makeContactCommand(cl *apiclient.APIClient, cmd utils.CommandType, contact Contact, oldContact Contact, diag diag.Diagnostics) *response.Response {
	req := map[string]interface{}{
		"COMMAND": fmt.Sprintf("%sContact", cmd),
	}

	if cmd == utils.CommandCreate {
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

	if cmd == utils.CommandCreate || cmd == utils.CommandUpdate {
		optionals := []string{"TITLE", "MIDDLENAME", "ORGANIZATION", "STATE", "FAX"}

		req["TITLE"] = utils.AutoUnboxString(contact.Title, "")
		req["FIRSTNAME"] = utils.AutoUnboxString(contact.FirstName, "")
		req["MIDDLENAME"] = utils.AutoUnboxString(contact.MiddleName, "")
		req["LASTNAME"] = utils.AutoUnboxString(contact.LastName, "")

		req["ORGANIZATION"] = utils.AutoUnboxString(contact.Organization, "")

		req["STREET0"] = utils.AutoUnboxString(contact.AddressLine1, "")
		req["STREET1"] = utils.AutoUnboxString(contact.AddressLine2, "")
		req["CITY"] = utils.AutoUnboxString(contact.City, "")
		req["STATE"] = utils.AutoUnboxString(contact.State, "")
		req["ZIP"] = utils.AutoUnboxString(contact.ZIP, "")
		req["COUNTRY"] = utils.AutoUnboxString(contact.Coutry, "")

		req["PHONE"] = utils.AutoUnboxString(contact.Phone, "")
		req["FAX"] = utils.AutoUnboxString(contact.Fax, "")
		req["EMAIL"] = utils.AutoUnboxString(contact.Email, "")

		if !contact.Disclose.Null && !contact.Disclose.Unknown {
			req["DISCLOSE"] = utils.BoolToNumberStr(contact.Disclose.Value)
		}

		if cmd == utils.CommandUpdate {
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

		utils.HandleExtraAttributesWrite(contact.ExtraAttributes, oldContact.ExtraAttributes, req)
	}

	return cl.Request(req)
}

func kindContactRead(ctx context.Context, contact Contact, cl *apiclient.APIClient, diag diag.Diagnostics) Contact {
	resp := makeContactCommand(cl, utils.CommandRead, contact, contact, diag)
	if diag.HasError() {
		return Contact{}
	}

	return Contact{
		ID: types.String{Value: utils.ColumnFirstOrDefault(resp, "ID", "").(string)},

		Title:      utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "TITLE", nil)),
		FirstName:  utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "FIRSTNAME", nil)),
		MiddleName: utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "MIDDLENAME", nil)),
		LastName:   utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "LASTNAME", nil)),

		Organization: utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "ORGANIZATION", nil)),

		AddressLine1: utils.AutoBoxString(utils.ColumnIndexOrDefault(resp, "STREET", nil, 0)),
		AddressLine2: utils.AutoBoxString(utils.ColumnIndexOrDefault(resp, "STREET", nil, 1)),
		City:         utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "CITY", nil)),
		ZIP:          utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "ZIP", nil)),
		State:        utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "STATE", nil)),
		Coutry:       utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "COUNTRY", nil)),

		Phone: utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "PHONE", nil)),
		Fax:   utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "FAX", nil)),
		Email: utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "EMAIL", nil)),

		Disclose: utils.AutoBoxBoolNumberStr(utils.ColumnFirstOrDefault(resp, "DISCLOSE", "0")),

		ExtraAttributes: utils.HandleExtraAttributesRead(resp),
	}
}
