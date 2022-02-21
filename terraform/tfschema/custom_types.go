package tfschema

import (
	"context"
	fmt "fmt"

	apitypes "github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/types/wrappers"
	"github.com/gravitational/teleport/api/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GenSchemaBoolOptions returns Terraform schema for BoolOption type
func GenSchemaBoolOption(_ context.Context) tfsdk.Attribute {
	return tfsdk.Attribute{
		Optional: true,
		Type:     types.BoolType,
	}
}

// GenSchemaBoolOptions returns Terraform schema for Traits type
func GenSchemaTraits(_ context.Context) tfsdk.Attribute {
	return tfsdk.Attribute{
		Optional: true,
		Type: types.MapType{
			ElemType: types.ListType{
				ElemType: types.StringType,
			},
		},
	}
}

// GenSchemaBoolOptions returns Terraform schema for Labels type
func GenSchemaLabels(ctx context.Context) tfsdk.Attribute {
	return GenSchemaTraits(ctx)
}

func CopyFromBoolOption(diags diag.Diagnostics, tf attr.Value, o **apitypes.BoolOption) {
	v, ok := tf.(types.Bool)
	if !ok {
		diags.AddError("Error reading from Terraform object", fmt.Sprintf("Can not convert %T to types.Bool", tf))
		return
	}
	value := apitypes.BoolOption{Value: v.Value}
	*o = &value
}

func CopyToBoolOption(diags diag.Diagnostics, o *apitypes.BoolOption, t attr.Type, v attr.Value) attr.Value {
	value, ok := v.(types.Bool)
	if !ok {
		value = types.Bool{}
	}

	if o == nil {
		value.Null = true
		return value
	}

	value.Value = o.Value

	return value
}

func CopyFromLabels(diags diag.Diagnostics, v attr.Value, o *apitypes.Labels) {
	value, ok := v.(types.Map)
	if !ok {
		diags.AddError("Error reading from Terraform object", fmt.Sprintf("Can not convert %T to types.Map", v))
		return
	}

	*o = make(map[string]utils.Strings, len(value.Elems))
	for k, e := range value.Elems {
		l, ok := e.(types.List)
		if !ok {
			diags.AddError("Error reading from Terraform object", fmt.Sprintf("Can not convert %T to types.List", l))
			return
		}

		(*o)[k] = make(utils.Strings, len(l.Elems))

		for i, v := range l.Elems {
			s, ok := v.(types.String)
			if !ok {
				diags.AddError("Error reading from Terraform object", fmt.Sprintf("Can not convert %T to types.String", s))
				return
			}

			(*o)[k][i] = s.Value
		}
	}
}

func CopyToLabels(diags diag.Diagnostics, o apitypes.Labels, t attr.Type, v attr.Value) attr.Value {
	typ := t.(types.MapType) // By the convention, t comes type-asserted so there would be no failure

	value, ok := v.(types.Map)
	if !ok {
		value = types.Map{ElemType: typ.ElemType}
	}

	if value.Elems == nil {
		value.Elems = make(map[string]attr.Value, len(o))
	}

	for k, l := range o {
		row := types.List{
			ElemType: types.StringType,
			Elems:    make([]attr.Value, len(l)),
		}

		for i, e := range l {
			row.Elems[i] = types.String{Value: e}
		}

		value.Elems[k] = row
	}

	return value
}

func CopyFromTraits(diags diag.Diagnostics, v attr.Value, o *wrappers.Traits) {
	//CopyFromLabels(diags, tf, o)
}

func CopyToTraits(diags diag.Diagnostics, o wrappers.Traits, t attr.Type, v attr.Value) attr.Value {
	return types.Map{Null: true, ElemType: types.ListType{ElemType: types.StringType}}
}

type CustomType interface {
	ToTerraformValue()
	FromTerraformValue()
}
