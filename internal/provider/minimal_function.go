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
var _ function.Function = &MinimalFunction{}

type MinimalFunction struct{}

func NewMinimalFunction() function.Function {
	return &MinimalFunction{}
}

func (f *MinimalFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "minimal"
}

func (f *MinimalFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Remove null, empty string, and empty collection values from a list or map",
		Description: "Returns the input list or map with all null values, empty strings, and empty collections (lists, maps, objects) removed.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "The list or map to minimize.",
			},
		},
		Return: function.DynamicReturn{},
	}
}

func isEmptyString(v attr.Value) bool {
	s, ok := v.(types.String)
	return ok && !s.IsNull() && s.ValueString() == ""
}

func isEmptyValue(v attr.Value) bool {
	switch c := v.(type) {
	case types.List:
		return len(c.Elements()) == 0
	case types.Tuple:
		return len(c.Elements()) == 0
	case types.Map:
		return len(c.Elements()) == 0
	case types.Object:
		return len(c.Attributes()) == 0
	}
	return false
}

func (f *MinimalFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))
	if resp.Error != nil {
		return
	}

	var result attr.Value

	notNullOrEmpty := func(v attr.Value) bool { return !v.IsNull() && !isEmptyString(v) && !isEmptyValue(v) }

	switch v := input.UnderlyingValue().(type) {
	case types.List:
		list, diags := types.ListValue(v.ElementType(ctx), filterListByPredicate(v.Elements(), notNullOrEmpty))
		if diags.HasError() {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
			return
		}
		result = list

	case types.Map:
		m, diags := types.MapValue(v.ElementType(ctx), filterMapByPredicate(v.Elements(), notNullOrEmpty))
		if diags.HasError() {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
			return
		}
		result = m

	case types.Object:
		filteredVals := filterMapByPredicate(v.Attributes(), notNullOrEmpty)
		filteredTypes := make(map[string]attr.Type, len(filteredVals))
		for k, val := range filteredVals {
			filteredTypes[k] = val.Type(ctx)
		}
		obj, diags := types.ObjectValue(filteredTypes, filteredVals)
		if diags.HasError() {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
			return
		}
		result = obj

	default:
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("input must be a list, map, or object"))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(result)))
}
