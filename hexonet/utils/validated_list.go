package utils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type validatedListType struct {
	elemType xattr.TypeWithValidate
	listType types.ListType
}

func ValidatedListType(elemType xattr.TypeWithValidate) *validatedListType {
	return &validatedListType{
		elemType: elemType,
		listType: types.ListType{
			ElemType: elemType,
		},
	}
}

var _ xattr.TypeWithValidate = &validatedListType{}

func (t *validatedListType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return t.listType.ApplyTerraform5AttributePathStep(step)
}

func (t *validatedListType) Equal(typ attr.Type) bool {
	other, ok := typ.(*validatedListType)
	if !ok {
		return false
	}
	return t.elemType.Equal(other.elemType)
}

func (t *validatedListType) String() string {
	return fmt.Sprintf("ValidatedList%s", t.elemType.String())
}

func (t *validatedListType) TerraformType(ctx context.Context) tftypes.Type {
	return t.listType.TerraformType(ctx)
}

func (t *validatedListType) ValueFromTerraform(ctx context.Context, val tftypes.Value) (attr.Value, error) {
	list, err := t.listType.ValueFromTerraform(ctx, val)
	if err != nil {
		return nil, err
	}

	return &ValidatedList{List: list.(types.List), attrType: t}, nil
}

func (t *validatedListType) Validate(ctx context.Context, val tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !val.Type().Equal(tftypes.List{
		ElementType: t.elemType.TerraformType(ctx),
	}) {
		diags.AddAttributeError(
			path,
			"Validated List Type Validation Error",
			fmt.Sprintf("Expected List value, received %T with value: %v", val, val),
		)
	}

	if !val.IsKnown() || val.IsNull() {
		return diags
	}

	data := make([]tftypes.Value, 0)
	err := val.As(&data)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Validated List Type Validation Error",
			fmt.Sprintf("Cannot convert value to list: %v", val),
		)
		return diags
	}

	for k, val := range data {
		diags = append(diags, t.elemType.Validate(ctx, val, path.AtListIndex(k))...)
	}

	return diags
}

func (t *validatedListType) NewList() *ValidatedList {
	return &ValidatedList{
		List: types.List{
			ElemType: t.elemType,
			Elems:    make([]attr.Value, 0),
		},
		attrType: t,
	}
}

type ValidatedList struct {
	types.List
	attrType *validatedListType
}

var _ attr.Value = &ValidatedList{}

func (l *ValidatedList) Type(ctx context.Context) attr.Type {
	return l.attrType
}
