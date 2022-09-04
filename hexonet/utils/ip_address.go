package utils

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ipAddressType struct {
	AllowIPv4 bool
	AllowIPv6 bool
}

func IPAddressType(allowIPv4 bool, allowIPv6 bool) *ipAddressType {
	if !allowIPv4 && !allowIPv6 {
		panic(errors.New("must set at least one of allowIPv4 or allowIPv6"))
	}

	return &ipAddressType{
		AllowIPv4: allowIPv4,
		AllowIPv6: allowIPv6,
	}
}

var _ xattr.TypeWithValidate = &ipAddressType{}

func (t *ipAddressType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return types.StringType.ApplyTerraform5AttributePathStep(step)
}

func (t *ipAddressType) Equal(typ attr.Type) bool {
	other, ok := typ.(*ipAddressType)
	if !ok {
		return false
	}
	return other.AllowIPv4 == t.AllowIPv4 && other.AllowIPv6 == t.AllowIPv6
}

func (t *ipAddressType) String() string {
	if t.AllowIPv4 {
		if t.AllowIPv6 {
			return "IPAddress"
		}
		return "IPv4Address"
	}
	// t.AllowIPv6 must be true here due to check in "IPAddressType(...)"
	return "IPv6Address"
}

func (t *ipAddressType) TerraformType(ctx context.Context) tftypes.Type {
	return types.StringType.TerraformType(ctx)
}

func (t *ipAddressType) ValueFromTerraform(ctx context.Context, val tftypes.Value) (attr.Value, error) {
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

	return t.IPFromString(s)
}

func (t *ipAddressType) Validate(ctx context.Context, val tftypes.Value, path path.Path) diag.Diagnostics {
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

	_, err = t.IPFromString(s)
	if err != nil {
		diags.AddAttributeError(
			path,
			"IP Address Type Validation Error",
			err.Error(),
		)
	}
	return diags
}

func (t *ipAddressType) IPFromString(s string) (ipAddress, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return ipAddress{}, fmt.Errorf("value is not a valid IP address: %s", s)
	}

	switch len(ip) {
	case net.IPv4len:
		if t.AllowIPv4 {
			break
		}
		return ipAddress{}, fmt.Errorf("value is an IPv4 address, which is not allowed: %s", s)
	case net.IPv6len:
		if t.AllowIPv6 {
			break
		}
		return ipAddress{}, fmt.Errorf("value is an IPv6 address, which is not allowed: %s", s)
	default:
		return ipAddress{}, fmt.Errorf("value is not a known type of IP address: %s", s)
	}

	return ipAddress{
		attrType: t,
		Unknown:  false,
		Null:     false,
		Value:    ip,
	}, nil
}

type ipAddress struct {
	attrType *ipAddressType
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
