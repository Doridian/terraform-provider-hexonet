package utils

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Supress diff if entire value is planned as "null"
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

// Suppress diff for any map entry planned as "null"
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

// Supress diff if IP addresses are canonically equal
func NoDiffIfEquivalentIP() tfsdk.AttributePlanModifier {
	return noDiffIfEquivalentIPModifier{}
}

type noDiffIfEquivalentIPModifier struct{}

func (r noDiffIfEquivalentIPModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	if req.AttributeConfig.IsUnknown() || req.AttributeConfig.IsNull() {
		return
	}

	if resp.AttributePlan.IsUnknown() || resp.AttributePlan.IsNull() {
		return
	}

	if req.AttributeState.IsUnknown() || req.AttributeState.IsNull() {
		return
	}

	srcIP := net.IP(req.AttributeState.(types.String).Value)
	destIP := net.IP(resp.AttributePlan.(types.String).Value)

	if srcIP == nil || destIP == nil {
		return
	}

	if srcIP.Equal(destIP) {
		resp.AttributePlan = req.AttributeState
	}
}

func (r noDiffIfEquivalentIPModifier) Description(ctx context.Context) string {
	return "Will suppress any diff if the target value is equivalent to the source value as an IP."
}

func (r noDiffIfEquivalentIPModifier) MarkdownDescription(ctx context.Context) string {
	return "Will suppress any diff if the target value is equivalent to the source value as an IP."
}

// Supress diff if IP addresses are canonically equal in a list
func NoDiffIfEquivalentIPList() tfsdk.AttributePlanModifier {
	return noDiffIfEquivalentIPListModifier{}
}

type noDiffIfEquivalentIPListModifier struct{}

func (r noDiffIfEquivalentIPListModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	if req.AttributeConfig.IsUnknown() || req.AttributeConfig.IsNull() {
		return
	}

	if resp.AttributePlan.IsUnknown() || resp.AttributePlan.IsNull() {
		return
	}

	if req.AttributeState.IsUnknown() || req.AttributeState.IsNull() {
		return
	}

	srcIPs := req.AttributeState.(types.List)
	destIPs := resp.AttributePlan.(types.List)

	for idx, ipBox := range destIPs.Elems {
		if idx >= len(srcIPs.Elems) {
			break
		}

		srcIP := net.IP(ipBox.(types.String).Value)
		destIP := net.IP(srcIPs.Elems[idx].(types.String).Value)

		if srcIP == nil || destIP == nil {
			continue
		}

		if srcIP.Equal(destIP) {
			destIPs.Elems[idx] = srcIPs.Elems[idx]
		}
	}
}

func (r noDiffIfEquivalentIPListModifier) Description(ctx context.Context) string {
	return "Will suppress any diff if the target value is equivalent to the source value as an IP."
}

func (r noDiffIfEquivalentIPListModifier) MarkdownDescription(ctx context.Context) string {
	return "Will suppress any diff if the target value is equivalent to the source value as an IP."
}
