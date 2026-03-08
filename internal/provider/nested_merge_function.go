// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = &NestedMergeFunction{}

type NestedMergeFunction struct{}

func NewNestedMergeFunction() function.Function {
	return &NestedMergeFunction{}
}

func (f *NestedMergeFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "nestedMerge"
}

func (f *NestedMergeFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Deep merge objects, recursively combining nested objects",
		Description: "Merges two or more objects left-to-right. When the same key exists in multiple arguments and both values are objects, they are recursively merged. For all other value types the rightmost argument wins, matching the behaviour of Terraform's built-in merge().",
		VariadicParameter: &function.DynamicParameter{
			Name:        "map",
			Description: "Objects or maps to merge. Later arguments take precedence for scalar values; objects are merged recursively.",
		},
		Return: function.DynamicReturn{},
	}
}

func (f *NestedMergeFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var maps []types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &maps))
	if resp.Error != nil {
		return
	}

	if len(maps) == 0 {
		empty, diags := types.ObjectValue(map[string]attr.Type{}, map[string]attr.Value{})
		if diags.HasError() {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
			return
		}
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(empty)))
		return
	}

	result := maps[0].UnderlyingValue()
	for _, next := range maps[1:] {
		merged, funcErr := nestedMergeTwo(ctx, result, next.UnderlyingValue())
		if funcErr != nil {
			resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
			return
		}
		result = merged
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(result)))
}

// nestedMergeTwo merges override into base. When both values are objects the
// merge is recursive; in all other cases override wins.
func nestedMergeTwo(ctx context.Context, base, override attr.Value) (attr.Value, *function.FuncError) {
	baseObj, baseIsObj := base.(types.Object)
	overrideObj, overrideIsObj := override.(types.Object)

	if baseIsObj && overrideIsObj {
		merged := make(map[string]attr.Value, len(baseObj.Attributes()))
		for k, v := range baseObj.Attributes() {
			merged[k] = v
		}
		for k, overrideVal := range overrideObj.Attributes() {
			if baseVal, exists := merged[k]; exists {
				mergedVal, funcErr := nestedMergeTwo(ctx, baseVal, overrideVal)
				if funcErr != nil {
					return nil, funcErr
				}
				merged[k] = mergedVal
			} else {
				merged[k] = overrideVal
			}
		}
		mergedTypes := make(map[string]attr.Type, len(merged))
		for k, v := range merged {
			mergedTypes[k] = v.Type(ctx)
		}
		obj, diags := types.ObjectValue(mergedTypes, merged)
		if diags.HasError() {
			return nil, function.FuncErrorFromDiags(ctx, diags)
		}
		return obj, nil
	}

	return override, nil
}
