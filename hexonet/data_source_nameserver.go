package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNameserver() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNameserverRead,
		Schema:      makeNameserverSchema(true),
	}
}

func dataSourceNameserverRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return kindNameserverRead(ctx, d, m, true)
}
