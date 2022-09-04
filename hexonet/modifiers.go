package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NoDiffIfNull() tfsdk.AttributePlanModifier {
	return noDiffIfNullModifier{}
}

type noDiffIfNullModifier struct{}

func (r noDiffIfNullModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	if req.AttributeConfig.IsUnknown() {
		return
	}

	if resp.AttributePlan.IsNull() && !resp.AttributePlan.IsUnknown() && !req.AttributeState.IsNull() && !req.AttributeState.IsUnknown() {
		resp.AttributePlan = req.AttributeState
	}
}

func (r noDiffIfNullModifier) Description(ctx context.Context) string {
	return "Will suppress any diff if the target value is null."
}

func (r noDiffIfNullModifier) MarkdownDescription(ctx context.Context) string {
	return "Will suppress any diff if the target value is null."
}

func NoDiffIfNullMapEntries() tfsdk.AttributePlanModifier {
	return noDiffIfNullMapEntriesModifier{}
}

type noDiffIfNullMapEntriesModifier struct{}

func (r noDiffIfNullMapEntriesModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	if req.AttributeState.IsNull() || req.AttributeState.IsUnknown() || resp.AttributePlan.IsNull() || resp.AttributePlan.IsUnknown() {
		return
	}

	if req.AttributeConfig.IsUnknown() {
		return
	}

	state := req.AttributeState.(types.Map)
	plan := resp.AttributePlan.(types.Map)

	for k, stateV := range state.Elems {
		if stateV.IsNull() || stateV.IsUnknown() {
			continue
		}

		planV, ok := plan.Elems[k]
		if ok && (!planV.IsNull() || planV.IsUnknown()) {
			continue
		}

		plan.Elems[k] = stateV
	}
}

func (r noDiffIfNullMapEntriesModifier) Description(ctx context.Context) string {
	return "Will suppress any diff for map keys where the target value is null."
}

func (r noDiffIfNullMapEntriesModifier) MarkdownDescription(ctx context.Context) string {
	return "Will suppress any diff for map keys where the target value is null."
}
