package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceContact() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceContactRead,
		Schema:      makeContactSchema(true),
	}
}

func dataSourceContactRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return kindContactRead(ctx, d, m, true)
}
