// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &FilterFunction{}

type FilterFunction struct{}

func NewFilterFunction() function.Function {
	return &FilterFunction{}
}

func (f *FilterFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "filter"
}

func (f *FilterFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Filter a map or object by value",
		Description: "Returns entries from the input map or object where attributes match any of the given condition objects. Within a single condition object all attributes must match (AND). Multiple condition objects are OR'd together. Nested condition values are matched as subsets: an object condition value matches if all its attributes are present in the target.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "The map or object to filter.",
			},
		},
		VariadicParameter: &function.DynamicParameter{
			Name:           "condition",
			Description:    "One or more condition objects. An entry is kept if it matches any one condition object (OR). Within each object all key=value pairs must match (AND). Object values are matched as subsets, enabling nested attribute filtering.",
			AllowNullValue: true,
		},
		Return: function.DynamicReturn{},
	}
}

func parseConditions(conditions []types.Dynamic) ([]map[string]attr.Value, *function.FuncError) {
	if len(conditions) == 0 {
		return nil, function.NewFuncError("at least one condition object is required")
	}
	var result []map[string]attr.Value
	for _, cond := range conditions {
		if cond.IsNull() || cond.IsUnknown() {
			continue
		}
		obj, ok := cond.UnderlyingValue().(types.Object)
		if !ok {
			return nil, function.NewFuncError("each condition must be an object")
		}
		result = append(result, obj.Attributes())
	}
	if len(result) == 0 {
		return nil, function.NewFuncError("at least one non-null condition object is required")
	}
	return result, nil
}

// attrMatchesConditionValue returns true if attrVal satisfies condVal.
// When condVal is an object, it performs a subset match: every attribute in
// condVal must exist in attrVal and itself satisfy the condition (recursive).
// For all other types, exact equality is required.
func attrMatchesConditionValue(attrVal, condVal attr.Value) bool {
	condObj, condIsObj := condVal.(types.Object)
	attrObj, attrIsObj := attrVal.(types.Object)
	if condIsObj {
		if !attrIsObj {
			return false
		}
		attrAttrs := attrObj.Attributes()
		for k, cv := range condObj.Attributes() {
			av, exists := attrAttrs[k]
			if !exists || !attrMatchesConditionValue(av, cv) {
				return false
			}
		}
		return true
	}
	return attrVal.Equal(condVal)
}

func matchesAnyCondition(v attr.Value, conditions []map[string]attr.Value) bool {
	obj, ok := v.(types.Object)
	if !ok {
		return false
	}
	attrs := obj.Attributes()
	for _, cond := range conditions {
		allMatch := true
		for condKey, condVal := range cond {
			attrVal, exists := attrs[condKey]
			if !exists || !attrMatchesConditionValue(attrVal, condVal) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return true
		}
	}
	return false
}

func filterListByPredicate(elems []attr.Value, keep func(attr.Value) bool) []attr.Value {
	filtered := make([]attr.Value, 0, len(elems))
	for _, v := range elems {
		if keep(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func filterMapByPredicate(elems map[string]attr.Value, keep func(attr.Value) bool) map[string]attr.Value {
	filtered := make(map[string]attr.Value, len(elems))
	for k, v := range elems {
		if keep(v) {
			filtered[k] = v
		}
	}
	return filtered
}

func filterAttrMap(elems map[string]attr.Value, conditions []map[string]attr.Value) map[string]attr.Value {
	return filterMapByPredicate(elems, func(v attr.Value) bool {
		return matchesAnyCondition(v, conditions)
	})
}

func (f *FilterFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic
	var conditionSlice []types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input, &conditionSlice))
	if resp.Error != nil {
		return
	}

	conditions, funcErr := parseConditions(conditionSlice)
	if funcErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
		return
	}

	var result attr.Value

	switch v := input.UnderlyingValue().(type) {
	case types.Map:
		filtered := filterAttrMap(v.Elements(), conditions)
		m, diags := types.MapValue(v.ElementType(ctx), filtered)
		if diags.HasError() {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
			return
		}
		result = m

	case types.Object:
		filtered := filterAttrMap(v.Attributes(), conditions)
		filteredTypes := make(map[string]attr.Type, len(filtered))
		for k, val := range filtered {
			filteredTypes[k] = val.Type(ctx)
		}
		obj, diags := types.ObjectValue(filteredTypes, filtered)
		if diags.HasError() {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
			return
		}
		result = obj

	default:
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("input must be a map or object"))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(result)))
}
