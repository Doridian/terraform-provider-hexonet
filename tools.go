//go:build tools

package main

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs" // Needed so go:generate below works
)
