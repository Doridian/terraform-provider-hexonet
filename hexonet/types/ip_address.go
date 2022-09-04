package types

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type IPAddressType struct{}

var _ xattr.TypeWithValidate = &IPAddressType{}

func (t *IPAddressType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return types.StringType.ApplyTerraform5AttributePathStep(step)
}

func (t *IPAddressType) Equal(typ attr.Type) bool {
	_, ok := typ.(*IPAddressType)
	return ok
}

func (t *IPAddressType) String() string {
	return "IPAddress"
}

func (t *IPAddressType) TerraformType(ctx context.Context) tftypes.Type {
	return types.StringType.TerraformType(ctx)
}

func (t *IPAddressType) ValueFromTerraform(ctx context.Context, val tftypes.Value) (attr.Value, error) {
	if !val.IsKnown() {
		return ipAddress{attrType: t, Unknown: true}, nil
	}

	if val.IsNull() {
		return ipAddress{attrType: t, Null: true}, nil
	}

	var s string
	err := val.As(&s)
	if err != nil {
		return nil, err
	}

	return ipAddress{attrType: t, Value: net.ParseIP(s)}, nil
}

func (t *IPAddressType) Validate(ctx context.Context, val tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !val.Type().Equal(tftypes.String) {
		diags.AddAttributeError(
			path,
			"IP Address Type Validation Error",
			fmt.Sprintf("Expected String value, received %T with value: %v", val, val),
		)
	}

	if !val.IsKnown() || val.IsNull() {
		return diags
	}

	var s string
	err := val.As(&s)
	if err != nil {
		diags.AddAttributeError(
			path,
			"IP Address Type Validation Error",
			fmt.Sprintf("Cannot convert value to string: %s", err),
		)
		return diags
	}

	ip := net.ParseIP(s)
	if ip == nil {
		diags.AddAttributeError(
			path,
			"IP Address Type Validation Error",
			fmt.Sprintf("Value is not a valid IP address: %s", s),
		)
		return diags
	}

	return diags
}

type ipAddress struct {
	attrType *IPAddressType
	Unknown  bool
	Null     bool
	Value    net.IP
}

var _ attr.Value = ipAddress{}

func (ip ipAddress) Equal(other attr.Value) bool {
	otherIP, ok := other.(ipAddress)
	if !ok {
		return false
	}

	if ip.Unknown || otherIP.Unknown {
		return false
	}

	if ip.Null {
		return otherIP.Null
	}

	return otherIP.Value.Equal(ip.Value)
}

func (ip ipAddress) IsNull() bool {
	return ip.Null
}

func (ip ipAddress) IsUnknown() bool {
	return ip.Unknown
}

func (ip ipAddress) String() string {
	return ip.Value.String()
}

func (ip ipAddress) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	if ip.Null {
		return tftypes.NewValue(tftypes.String, nil), nil
	}

	if ip.Unknown {
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), nil
	}

	return tftypes.NewValue(tftypes.String, ip.Value.String()), nil
}

func (ip ipAddress) Type(ctx context.Context) attr.Type {
	return ip.attrType
}
